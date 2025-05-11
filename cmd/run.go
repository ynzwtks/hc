package cmd

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

// runCmd Run the test set
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the test set",
	Long:  `Run the test set`,
	Run: func(cmd *cobra.Command, args []string) {
		commonInit()

		runtimeInit()

		//テストID指定(-t)の場合
		if opt.target != -1 {
			if opt.target >= set.TestDataNum || opt.target < 0 {
				errorPrint("The test number is out of range")
				return
			}
			runSingleCmd(fmt.Sprintf("%04d", ri.testID[0]))
			return
		}

		if !opt.quietMode {
			printTitle()
		}
		workerPool()

		//printLargeScore(5)
		if ri.enableLog {
			printLog()
		}
	},
}

func printLog() {
	sort.Slice(ri.score, func(i, j int) bool {
		return ri.score[i].a < ri.score[j].a
	})

	fs := make([]string, 0)
	for i := 0; i < len(ri.score); i++ {
		if ri.score[i].a >= 0 && ri.score[i].b == 0 {
			ri.score[i].b = -1
			fs = append(fs, fmt.Sprintf("%04d", ri.score[i].a))
		}

	}

	tot, ave, cntOk, cntNg := 0, 0, 0, 0
	totLog := 0.0
	aveLog := 0
	for i := 0; i < set.TestDataNum; i++ {
		if ri.score[i].b > 0 {
			cntOk++
			tot += ri.score[i].b
			totLog += math.Log(float64(ri.score[i].b))
		} else {
			cntNg++
		}
	}
	if cntOk > 0 {
		ave = tot / cntOk
		aveLog = int(math.Round(math.Exp(totLog / float64(cntOk))))
	}
	// Optput
	// lipglossスタイルの定義
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	dataStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("white"))

	// ヘッダー
	header := fmt.Sprintf("%s%s%s%s%s", headerStyle.Width(20).Align(lipgloss.Left).Render("Date"),
		headerStyle.Width(20).Align(lipgloss.Left).Render("Geometric Mean"),
		headerStyle.Width(20).Align(lipgloss.Left).Render("Arithmetic Mean"),
		headerStyle.Width(20).Align(lipgloss.Left).Render("Test Case Count"),
		headerStyle.Width(20).Align(lipgloss.Left).Render("Error Count"))
	// データ
	data := fmt.Sprintf("%s%s%s%s%s", dataStyle.Width(20).Align(lipgloss.Left).Render(stringTime()),
		dataStyle.Width(20).Align(lipgloss.Center).Render(fmt.Sprintf("%d", aveLog)),
		dataStyle.Width(20).Align(lipgloss.Center).Render(fmt.Sprintf("%d", ave)),
		dataStyle.Width(20).Align(lipgloss.Center).Render(fmt.Sprintf("%d", set.TestDataNum)),
		dataStyle.Width(20).Align(lipgloss.Center).Render(fmt.Sprintf("%d", cntNg)))

	//fs
	// 出力
	if opt.quietMode {
		fmt.Println(aveLog, ave, set.TestDataNum, cntNg)
	} else {
		fmt.Println("")
		fmt.Println(header)
		fmt.Println(data)
		if len(fs) > 0 {
			fmt.Println("")
			fmt.Println("Error Cases")
			fmt.Println(dataStyle.Width(80).Align(lipgloss.Left).Render(fmt.Sprintln(fs)))
		}
	}

	now := stringTime()
	if ri.enableLog {
		runCsv := fmt.Sprintf("%s/%s", logs.logDir, RunCsv)
		logLine := fmt.Sprintf("%s,%s,%d,%d,%d,%d,%d\n", now, opt.logMsg, cntOk, cntNg, tot, aveLog, ave)
		writeToFile(runCsv, []byte(logLine), true)
	}
	if ri.enableLogStandings {
		d := make([]int, 0)
		for i := 0; i < set.TestDataNum; i++ {
			d = append(d, ri.score[i].b)
		}
		historyCsv := fmt.Sprintf("%s/%s", logs.logDir, HistoryCsv)
		counter := int64(0)
		dat := intsToCsv(d)
		if countLines(historyCsv) > 0 {
			line := readLastLine(historyCsv)
			lines := strings.Split(line, ",")
			counter, _ = strconv.ParseInt(lines[1], 10, 64)
			counter++
		}
		head := fmt.Sprintf("%s,%04d,%s,%s\n", now, counter, opt.logMsg, dat)
		writeToFile(historyCsv, []byte(head), true)
		if sd.Enable == true {
			resultCSV := fmt.Sprintf("%s/%s", logs.logDir, ResultCsv)
			l := fmt.Sprintf("%04d:%s,%s", counter, opt.logMsg, dat)
			insertLine(resultCSV, 2, l)
		}
	}
}

func runtimeInit() {
	ri.score = make([]pair, set.TestDataNum)

	if len(opt.filter) == 0 && opt.loop == 1 {
		ri.enableLog = true
	} else {
		ri.enableLog = false
	}

	if ri.enableLog && len(opt.logMsg) > 0 {
		ri.enableLogStandings = true
	}

	ri.lastDist = make([]int, 20)
	ri.bestDist = make([]int, 20)
}

func workerPool() {
	ri.score = make([]pair, set.TestDataNum)
	ri.scoreSum = 0
	ri.ng = make([]int, 0)

	var mutex sync.Mutex
	// ワーカーの数
	numWorkers := cmn.Workers
	ri.executingCase = make([]string, numWorkers)
	// 実行するコマンドの総数
	numCommands := len(ri.testID) * int(opt.loop)
	//ri.lastDisplayTime = time.Now()
	// タスク用のチャネル
	tasks := make(chan string, numCommands)

	// WaitGroupの初期化
	var wg sync.WaitGroup

	//プログレスバー
	//progressbar.DefaultSilent()
	ri.bar = progressbar.NewOptions(numCommands,
		progressbar.OptionSetWidth(50), // プログレスバーの長さを50に設定
		progressbar.OptionShowCount(),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWriter(io.Discard))

	if !opt.quietMode {
		draw(&mutex)
	}
	// ワーカーの起動
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, &wg, tasks, &mutex)
	}
	for l := 1; l <= int(opt.loop); l++ {
		for i := 0; i < len(ri.testID); i++ {
			taskId := fmt.Sprintf("%04d %d", ri.testID[i], l)
			tasks <- taskId
		}
	}

	// 全てのタスクを追加したらチャネルを閉じる
	close(tasks)
	// 全ワーカーの処理完了を待つ
	wg.Wait()
	if !opt.quietMode {
		ri.lastDisplayTime = time.Now().Add(-24 * time.Hour)
		draw(&mutex)
	}
}

func worker(id int, wg *sync.WaitGroup, tasks <-chan string, mutex *sync.Mutex) {
	defer wg.Done()
	for task := range tasks {
		task = strings.Fields(task)[0]
		ri.executingCase[id] = task
		idx, _ := strconv.Atoi(task)

		sc, ok := runTestCmd(task)
		if ok == false {
			ri.ng = append(ri.ng, idx)
		}
		mutex.Lock()

		ri.scoreSum += sc
		if sc != 0 {
			ri.scoreLogSum += math.Log(float64(sc))
		}
		ri.executed++
		ri.bar.Add(1)
		mutex.Unlock()
		ri.score[idx].a = idx
		ri.score[idx].b = max(ri.score[idx].b, sc)

		if ri.score[idx].b == 0 {
			ri.score[idx].b = sc
		} else {
			if cmn.IsRankMin {
				ri.score[idx].b = max(ri.score[idx].b, sc)
			} else {
				ri.score[idx].b = min(ri.score[idx].b, sc)
			}
		}

		tid, _ := strconv.Atoi(task)
		if !opt.quietMode {
			applyResult(id, tid, task, mutex)
			draw(mutex)
		}
		ri.executingCase[id] = ""
	}
}
func applyResult(id int, tid int, task string, mutex *sync.Mutex) {
	mutex.Lock()

	if ri.score[tid].b <= 0 {
		ri.ngCnt++
		ri.failedTask = append(ri.failedTask, task)
	} else {
		var f1, f2 float64
		ri.okCnt++
		if logs.last[tid] > 0 {
			f1 = ((float64(ri.score[tid].b) - float64(logs.last[tid])) / float64(logs.last[tid])) * 100
			if logs.last[tid] != INF {
				if f1 > 0 {
					ri.incLast = append(ri.incLast, scoreElem{ratio: f1, id: task, oldScore: logs.last[tid], newScore: ri.score[tid].b})
				} else if f1 < 0 {
					ri.decLast = append(ri.decLast, scoreElem{ratio: f1, id: task, oldScore: logs.last[tid], newScore: ri.score[tid].b})
				}

			}
		}
		if set.IsSystemTest {
			if logs.best2[tid] > 0 {
				f2 = ((float64(ri.score[tid].b) - float64(logs.best2[tid])) / float64(logs.best2[tid])) * 100
				if float64(logs.best2[tid]) != INF {
					if f2 > 0 {
						ri.incBest = append(ri.incBest, scoreElem{ratio: f2, id: task, oldScore: logs.best2[tid], newScore: ri.score[tid].b})
					} else if f2 < 0 {
						ri.decBest = append(ri.decBest, scoreElem{ratio: f2, id: task, oldScore: logs.best2[tid], newScore: ri.score[tid].b})
					}
				}
			}
		} else {
			if logs.best[tid] > 0 {
				f2 = ((float64(ri.score[tid].b) - float64(logs.best[tid])) / float64(logs.best[tid])) * 100
				if float64(logs.best[tid]) != INF {
					if f2 > 0 {
						ri.incBest = append(ri.incBest, scoreElem{ratio: f2, id: task, oldScore: logs.best[tid], newScore: ri.score[tid].b})
					} else if f2 < 0 {
						ri.decBest = append(ri.decBest, scoreElem{ratio: f2, id: task, oldScore: logs.best[tid], newScore: ri.score[tid].b})
					}
				}
			}
		}

		sort.Slice(ri.decLast, func(i, j int) bool {
			return ri.decLast[i].ratio < ri.decLast[j].ratio
		})
		if len(ri.decLast) > 3 {
			ri.decLast = ri.decLast[:3]
		}

		sort.Slice(ri.incLast, func(i, j int) bool {
			return ri.incLast[i].ratio > ri.incLast[j].ratio
		})
		if len(ri.incLast) > 3 {
			ri.incLast = ri.incLast[:3]
		}

		sort.Slice(ri.decBest, func(i, j int) bool {
			return ri.decBest[i].ratio < ri.decBest[j].ratio
		})
		if len(ri.decBest) > 3 {
			ri.decBest = ri.decBest[:3]
		}

		sort.Slice(ri.incBest, func(i, j int) bool {
			return ri.incBest[i].ratio > ri.incBest[j].ratio
		})
		if len(ri.incBest) > 3 {
			ri.incBest = ri.incBest[:3]
		}

		if len(logs.last) != 0 && logs.last[tid] != INF {
			if f1 < -160 {
				ri.lastDist[0]++
			} else if f1 < -80 {
				ri.lastDist[1]++
			} else if f1 < -40 {
				ri.lastDist[2]++
			} else if f1 < -20 {
				ri.lastDist[3]++
			} else if f1 < -10 {
				ri.lastDist[4]++
			} else if f1 < 0 {
				ri.lastDist[5]++
			} else if f1 == 0 {
				ri.lastDist[6]++
			} else if f1 <= 10 {
				ri.lastDist[7]++
			} else if f1 <= 20 {
				ri.lastDist[8]++
			} else if f1 <= 40 {
				ri.lastDist[9]++
			} else if f1 <= 80 {
				ri.lastDist[10]++
			} else if f1 <= 160 {
				ri.lastDist[11]++
			} else {
				ri.lastDist[12]++
			}
		}

		if len(logs.best) != 0 && logs.best[tid] != INF {
			if f2 < -160 {
				ri.bestDist[0]++
			} else if f2 < -80 {
				ri.bestDist[1]++
			} else if f2 < -40 {
				ri.bestDist[2]++
			} else if f2 < -20 {
				ri.bestDist[3]++
			} else if f2 < -10 {
				ri.bestDist[4]++
			} else if f2 < 0.0 {
				ri.bestDist[5]++
			} else if f2 == 0 {
				ri.bestDist[6]++
			} else if f2 <= 10 {
				ri.bestDist[7]++
			} else if f2 <= 20 {
				ri.bestDist[8]++
			} else if f2 <= 40 {
				ri.bestDist[9]++
			} else if f2 <= 80 {
				ri.bestDist[10]++
			} else if f2 <= 160 {
				ri.bestDist[11]++
			} else if f2 > 160 {
				ri.bestDist[12]++
			}
		}

	}

	mutex.Unlock()
}
func draw(mutex *sync.Mutex) {

	mutex.Lock()
	if time.Since(ri.lastDisplayTime) <= 1000*time.Millisecond {
		mutex.Unlock()
		return
	}
	ri.lastDisplayTime = time.Now()

	printLineBack(14)
	fmt.Println(ri.bar.String())
	fmt.Println("")

	var fsl string
	var rsl string
	var asl string

	if len(ri.failedTask) > 5 {
		t := slices.Clone(ri.failedTask[len(ri.failedTask)-5:])
		sort.Strings(t)
		t = append(t, "...")
		fsl = fmt.Sprintf("%d %v", ri.ngCnt, ri.failedTask[len(ri.failedTask)-5:])
	} else {
		fsl = fmt.Sprintf("%d %v", ri.ngCnt, ri.failedTask)
	}
	ec := make([]string, 0)
	for i := 0; i < len(ri.executingCase); i++ {
		if len(ri.executingCase[i]) != 0 {
			ec = append(ec, ri.executingCase[i])
		}
	}
	sort.Strings(ec)
	if len(ec) > 9 {
		ec = ec[:9]
		ec = append(ec, "...")
	}
	rsl = fmt.Sprintf(" %v", ec)
	ave := 0
	aveLog := 0
	if ri.okCnt != 0 {
		ave = ri.scoreSum / ri.okCnt
		aveLog = int(math.Round(math.Exp(ri.scoreLogSum / float64(ri.okCnt))))
	}
	asl = fmt.Sprintf("%d(%d)", aveLog, ave)
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Width(40).Bold(true)
	fmt.Printf("%-40s%-40s%-40s\n", title.Render("Mean"), title.Render("Failed"), title.Render("Running"))
	fmt.Printf("%-40s%-40s%-40s\n", asl, fsl, rsl)
	fmt.Println("")
	sv := make([][]string, 4)
	var t string

	for i := 0; i < len(ri.decLast); i++ {
		t = fmt.Sprintf("[%s] %d->%d(%3.2f%%)", ri.decLast[i].id, ri.decLast[i].oldScore, ri.decLast[i].newScore, ri.decLast[i].ratio)
		sv[0] = append(sv[0], t)
	}
	for i := 0; i < len(ri.decBest); i++ {
		t = fmt.Sprintf("[%s] %d->%d(%3.2f%%)", ri.decBest[i].id, ri.decBest[i].oldScore, ri.decBest[i].newScore, ri.decBest[i].ratio)
		sv[1] = append(sv[1], t)
	}

	for i := 0; i < len(ri.incLast); i++ {
		t = fmt.Sprintf("[%s] %d->%d(%3.2f%%)", ri.incLast[i].id, ri.incLast[i].oldScore, ri.incLast[i].newScore, ri.incLast[i].ratio)
		sv[2] = append(sv[2], t)
	}
	for i := 0; i < len(ri.incBest); i++ {
		t = fmt.Sprintf("[%s] %d->%d(%3.2f%%)", ri.incBest[i].id, ri.incBest[i].oldScore, ri.incBest[i].newScore, ri.incBest[i].ratio)
		sv[3] = append(sv[3], t)
	}
	sp := lipgloss.NewStyle().Align(lipgloss.Left).Width
	wh := lipgloss.NewStyle().Bold(false).Width(10).Background(lipgloss.Color("white"))
	sc := lipgloss.NewStyle().Bold(false).Width(10).Align(lipgloss.Right).Background(lipgloss.Color("1")).Foreground(lipgloss.Color("8")).Bold(true)
	sc2 := lipgloss.NewStyle().Bold(false).Width(20).Align(lipgloss.Center).Background(lipgloss.Color("3")).Foreground(lipgloss.Color("8")).Bold(true)
	sc3 := lipgloss.NewStyle().Bold(false).Width(10).Align(lipgloss.Left).Background(lipgloss.Color("6")).Foreground(lipgloss.Color("8")).Bold(true)
	fmt.Println("")
	fmt.Printf("%s%s%s%s%s%s%s%s%s%s%s%s\n", sp(10).Render(""), sc.Render("-160%"), sc.Render("-80%"), sc.Render("-40"), sc.Render("-20%"), sc.Render("-10%"), sc2.Render("0%"), sc3.Render("10%"), sc3.Render("20%"), sc3.Render("40%"), sc3.Render("80%"), sc3.Render("160%"))
	fmt.Printf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s\n", wh.Render("Last"), sp(2).Render(""),
		sp(10).Render(itoa(ri.lastDist[0])), sp(10).Render(itoa(ri.lastDist[1])), sp(10).Render(itoa(ri.lastDist[2])),
		sp(10).Render(itoa(ri.lastDist[3])), sp(10).Render(itoa(ri.lastDist[4])), sp(7).Render(itoa(ri.lastDist[5])),
		sp(7).Render(itoa(ri.lastDist[6])), sp(10).Render(itoa(ri.lastDist[7])), sp(10).Render(itoa(ri.lastDist[8])),
		sp(10).Render(itoa(ri.lastDist[9])), sp(10).Render(itoa(ri.lastDist[10])), sp(10).Render(itoa(ri.lastDist[11])),
		sp(10).Render(itoa(ri.lastDist[12])))
	fmt.Printf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s\n", wh.Render("Best"), sp(2).Render(""),
		sp(10).Render(itoa(ri.bestDist[0])), sp(10).Render(itoa(ri.bestDist[1])), sp(10).Render(itoa(ri.bestDist[2])),
		sp(10).Render(itoa(ri.bestDist[3])), sp(10).Render(itoa(ri.bestDist[4])), sp(7).Render(itoa(ri.bestDist[5])),
		sp(7).Render(itoa(ri.bestDist[6])), sp(10).Render(itoa(ri.bestDist[7])), sp(10).Render(itoa(ri.bestDist[8])),
		sp(10).Render(itoa(ri.bestDist[9])), sp(10).Render(itoa(ri.bestDist[10])), sp(10).Render(itoa(ri.bestDist[11])),
		sp(10).Render(itoa(ri.bestDist[12])))
	fmt.Println("")

	top := lipgloss.NewStyle().Width(37).Foreground(lipgloss.Color("7")).Align(lipgloss.Left).Bold(true)
	top2 := lipgloss.NewStyle().Width(37).Foreground(lipgloss.Color("7")).Align(lipgloss.Left).Bold(true)
	fmt.Printf("%-37s%-37s%-37s%-37s\n", top.Render("Decrease(Last)"), top.Render("Decrease(Best)"), top2.Render("Increase(Last)"), top2.Render("Increase(Best)"))
	for i := 0; i < 3; i++ {
		t = ""
		for j := 0; j < 4; j++ {
			if len(sv[j]) > i {
				fmt.Printf("%-37s", sv[j][i])
			} else {
				fmt.Printf("%-37s", "-")
			}
		}
		fmt.Printf("\n")
	}

	mutex.Unlock()
}

func runTestCmd(id string) (int, bool) {
	if opt.debugMode {
		debugPrint("runTestCmd started with id=%s", id)
		debugPrint("TestDataPath=%s", set.TestDataPath)
	}

	// まず標準のテストファイルパスを試みる
	testFile := fmt.Sprintf("%s/%s.txt", set.TestDataPath, id)

	// 標準のパスが存在しない場合、"in"サブディレクトリを試みる
	if !fileExists(testFile) {
		inTestFile := fmt.Sprintf("%s/in/%s.txt", set.TestDataPath, id)
		if fileExists(inTestFile) {
			testFile = inTestFile
			if opt.debugMode {
				debugPrint("Using test file in 'in' subdirectory: %s", testFile)
			}
		}
	}

	if opt.debugMode {
		debugPrint("Looking for test file: %s", testFile)
		debugPrint("File exists: %v", fileExists(testFile))
	}

	// 入力ファイルが存在するか確認
	if !fileExists(testFile) {
		// 絶対パスを取得
		absPath, pathErr := filepath.Abs(testFile)
		if pathErr != nil {
			absPath = testFile // エラー時は相対パスだけを使用
		}
		errorPrint("Input file not found")
		fmt.Fprintf(os.Stderr, "  Absolute path: %s\n", absPath)
		return 0, false
	}

	envErr := setEnvVar("INPUT_FILE", testFile)
	if opt.debugMode && envErr != nil {
		debugPrint("Error setting INPUT_FILE environment variable: %v", envErr)
	}

	var s []string

	if cmn.IsInteractive == true {
		if opt.debugMode {
			debugPrint("Running interactive mode")
			debugPrint("JudgeProgram=%s", cmn.JudgeProgram)
			debugPrint("TargetProgram=%s", cmn.TargetProgram)
		}
		cmd := strings.Fields(cmn.JudgeProgram)
		cmd = append(cmd, strings.Fields(cmn.TargetProgram)...)
		if opt.debugMode {
			debugPrint("Full command=%v", cmd)
		}
		o1, o2, exitCode := ExecuteWithFileInput(testFile, cmd, false, false)
		if opt.debugMode {
			debugPrint("Interactive command exit code: %v", exitCode)
			if len(o1) > 0 {
				debugPrint("First 100 chars of stdout: %s", truncString(o1, 100))
			}
			if len(o2) > 0 {
				debugPrint("First 100 chars of stderr: %s", truncString(o2, 100))
			}
		}
		s = strings.Split(string(o2), "\n")
	} else {
		if opt.debugMode {
			debugPrint("Running non-interactive mode")
			debugPrint("TargetProgram=%s", cmn.TargetProgram)
		}
		cmd := strings.Fields(cmn.TargetProgram)
		if opt.debugMode {
			debugPrint("Target command=%v", cmd)
		}
		o1, o2, execErr := ExecuteWithFileInput(testFile, cmd, false, false)
		if opt.debugMode {
			if execErr != nil {
				debugPrint("Target command error: %v", execErr)
			}
			if len(o1) > 0 {
				debugPrint("First 100 chars of stdout: %s", truncString(o1, 100))
			}
			if len(o2) > 0 {
				debugPrint("First 100 chars of stderr: %s", truncString(o2, 100))
			}
		}

		tmpFile := fmt.Sprintf("%s/%s_o.txt", set.TestDataPath, id)
		if opt.debugMode {
			debugPrint("Writing output to file: %s", tmpFile)
		}
		writeErr := writeToFile(tmpFile, []byte(o1), false)
		if opt.debugMode && writeErr != nil {
			debugPrint("Error writing to output file: %v", writeErr)
		}

		if opt.debugMode {
			debugPrint("JudgeProgram=%s", cmn.JudgeProgram)
			debugPrint("Running judge command: %s %s %s", cmn.JudgeProgram, testFile, tmpFile)
		}
		o3, judgeErr := executeCommand([]string{cmn.JudgeProgram, testFile, tmpFile})
		if opt.debugMode {
			if judgeErr != nil {
				debugPrint("Judge command error: %v", judgeErr)
			}
			if len(o3) > 0 {
				debugPrint("First 100 chars of judge output: %s", truncString(string(o3), 100))
			}
		}
		s = strings.Split(string(o3), "\n")
	}

	lineIdx := -1
	for i := len(s) - 1; i >= 0; i-- {
		if strings.HasPrefix(s[i], cmn.ScoreLine) {
			lineIdx = i
			break
		}
	}

	if opt.debugMode {
		debugPrint("Looking for score line with prefix '%s'", cmn.ScoreLine)
		debugPrint("Score line index: %d", lineIdx)
	}

	// スコア行が見つからなかった場合のエラーハンドリング
	if lineIdx == -1 {
		warningPrint("Score line with prefix '%s' not found in output.", cmn.ScoreLine)
		if opt.debugMode {
			debugPrint("Output lines count: %d", len(s))
			if len(s) > 0 {
				debugPrint("Last few lines of output:")
				start := max(0, len(s)-5)
				for i := start; i < len(s); i++ {
					debugPrint("[%d] %s", i, s[i])
				}
			}
		}
		return 0, false
	}

	t := strings.Fields(s[lineIdx])
	if opt.debugMode {
		debugPrint("Score line: '%s'", s[lineIdx])
		debugPrint("Score fields: %v", t)
	}

	if len(t) == 0 {
		warningPrint("Empty score line found.")
		return 0, false
	}

	sc, parseErr := strconv.Atoi(t[len(t)-1])
	if opt.debugMode {
		if parseErr != nil {
			debugPrint("Error parsing score: %v", parseErr)
		} else {
			debugPrint("Parsed score: %d", sc)
		}
	}
	return sc, parseErr == nil
}
func runSingleCmd(id string) {
	if opt.debugMode {
		debugPrint("runSingleCmd started with id=%s", id)
		debugPrint("TestDataPath=%s", set.TestDataPath)
		wd, _ := os.Getwd()
		debugPrint("Current working directory: %s", wd)
	}

	// まず標準のテストファイルパスを試みる
	testFile := fmt.Sprintf("%s/%s.txt", set.TestDataPath, id)

	// 標準のパスが存在しない場合、"in"サブディレクトリを試みる
	if !fileExists(testFile) {
		inTestFile := fmt.Sprintf("%s/in/%s.txt", set.TestDataPath, id)
		if fileExists(inTestFile) {
			testFile = inTestFile
			if opt.debugMode {
				debugPrint("Using test file in 'in' subdirectory: %s", testFile)
			}
		}
	}

	if opt.debugMode {
		debugPrint("Looking for test file: %s", testFile)
		debugPrint("File exists: %v", fileExists(testFile))
	}
	// 入力ファイルが存在するか確認
	if !fileExists(testFile) {
		// 絶対パスを取得
		absPath, pathErr := filepath.Abs(testFile)
		if pathErr != nil {
			absPath = testFile // エラー時は相対パスだけを使用
		}

		errorPrint("Input file not found")
		fmt.Fprintf(os.Stderr, "  Relative path: %s\n", testFile)
		fmt.Fprintf(os.Stderr, "  Absolute path: %s\n", absPath)
		return
	}

	envErr := setEnvVar("INPUT_FILE", testFile)
	if opt.debugMode && envErr != nil {
		debugPrint("Error setting INPUT_FILE environment variable: %v", envErr)
	}

	var s []string
	var o1, o2 string
	var err error
	if cmn.IsInteractive == true {
		if opt.debugMode {
			debugPrint("Running interactive mode")
			debugPrint("JudgeProgram=%s", cmn.JudgeProgram)
			debugPrint("TargetProgram=%s", cmn.TargetProgram)
		}
		cmd := strings.Fields(cmn.JudgeProgram)
		cmd = append(cmd, strings.Fields(cmn.TargetProgram)...)
		if opt.debugMode {
			debugPrint("Full command=%v", cmd)
		}
		o1, o2, exitCode := ExecuteWithFileInput(testFile, cmd, false, false)
		if opt.debugMode {
			debugPrint("Interactive command exit code: %v", exitCode)
			if len(o1) > 0 {
				debugPrint("First 100 chars of stdout: %s", truncString(o1, 100))
			}
			if len(o2) > 0 {
				debugPrint("First 100 chars of stderr: %s", truncString(o2, 100))
			}
		}
		_ = o1
		s = strings.Split(string(o2), "\n")

	} else {
		if opt.debugMode {
			debugPrint("Running non-interactive mode")
			debugPrint("TargetProgram=%s", cmn.TargetProgram)
		}
		cmd := strings.Fields(cmn.TargetProgram)
		if opt.debugMode {
			debugPrint("Target command=%v", cmd)
		}
		o1, o2, err = ExecuteWithFileInput(testFile, cmd, false, true)
		if opt.debugMode {
			if err != nil {
				debugPrint("Target command error: %v", err)
			}
			if len(o1) > 0 {
				debugPrint("First 100 chars of stdout: %s", truncString(o1, 100))
			}
			if len(o2) > 0 {
				debugPrint("First 100 chars of stderr: %s", truncString(o2, 100))
			}
		}

		tmpFile := fmt.Sprintf("%s/out.txt", previousDirectory)
		if opt.debugMode {
			debugPrint("Writing output to file: %s", tmpFile)
		}
		writeErr := writeToFile(tmpFile, []byte(o1), false)
		if opt.debugMode && writeErr != nil {
			debugPrint("Error writing to output file: %v", writeErr)
		}

		if opt.debugMode {
			debugPrint("JudgeProgram=%s", cmn.JudgeProgram)
			debugPrint("Running judge command: %s %s %s", cmn.JudgeProgram, testFile, tmpFile)
		}
		o3, judgeErr := executeCommand([]string{cmn.JudgeProgram, testFile, tmpFile})
		if opt.debugMode {
			if judgeErr != nil {
				debugPrint("Judge command error: %v", judgeErr)
			}
			if len(o3) > 0 {
				debugPrint("First 100 chars of judge output: %s", truncString(string(o3), 100))
			}
		}
		s = strings.Split(string(o3), "\n")
	}
	lineIdx := -1
	for i := len(s) - 1; i >= 0; i-- {
		if strings.HasPrefix(s[i], cmn.ScoreLine) {
			lineIdx = i
			break
		}
	}

	if opt.debugMode {
		debugPrint("Looking for score line with prefix '%s'", cmn.ScoreLine)
		debugPrint("Score line index: %d", lineIdx)
	}

	// スコア行が見つからなかった場合のエラーハンドリング
	sc := 0
	var parseErr error
	if lineIdx == -1 {
		warningPrint("Score line with prefix '%s' not found in output.", cmn.ScoreLine)
		if opt.debugMode {
			debugPrint("Output lines count: %d", len(s))
			if len(s) > 0 {
				debugPrint("Last few lines of output:")
				start := max(0, len(s)-5)
				for i := start; i < len(s); i++ {
					debugPrint("[%d] %s", i, s[i])
				}
			}
		}
		parseErr = fmt.Errorf("score line not found")
	} else {
		t := strings.Fields(s[lineIdx])
		if opt.debugMode {
			debugPrint("Score line: '%s'", s[lineIdx])
			debugPrint("Score fields: %v", t)
		}
		if len(t) == 0 {
			warningPrint("Empty score line found.")
			parseErr = fmt.Errorf("empty score line")
		} else {
			sc, parseErr = strconv.Atoi(t[len(t)-1])
			if opt.debugMode {
				if parseErr != nil {
					debugPrint("Error parsing score: %v", parseErr)
				} else {
					debugPrint("Parsed score: %d", sc)
				}
			}
		}
	}

	idx, _ := strconv.Atoi(id)
	if parseErr != nil {
		ri.ng = append(ri.ng, idx)
		if opt.debugMode {
			debugPrint("Adding to NG list: %d", idx)
		}
	}

	if len(logs.vals2) != 0 && set.IsSystemTest {
		if opt.debugMode {
			debugPrint("System test mode, using best2 values")
		}
		fmt.Printf("No=%04d Score=%d  Best=%d Rank=%d/%d\n", opt.target, sc, logs.best2[opt.target], calcRank(sc, int(opt.target)), len(logs.vals2)+1)
	} else {
		if opt.debugMode {
			debugPrint("Normal mode, using best values")
		}
		fmt.Printf("No=%04d Score=%d  Best=%d Rank=%d/%d\n", opt.target, sc, logs.best[opt.target], calcRank(sc, int(opt.target)), len(logs.vals)+1)
	}

	if len(set.Seeds) != 0 {
		fmt.Printf("Parameter=%s Seed=%s\n", hi.HeaderData[opt.target], set.Seeds[opt.target])
	} else {
		fmt.Printf("Parameter=%s\n", hi.HeaderData[opt.target])
	}

	return
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&opt.setName, "set-name", "s", "default", "Set name to run")
	runCmd.Flags().BoolVarP(&opt.quietMode, "quiet", "q", false, "Run in quiet mode")
	runCmd.Flags().StringVarP(&opt.filter, "filter", "f", "", "Set filter definition")
	runCmd.Flags().IntVarP(&opt.loop, "loop", "l", 1, "Set filter definition")
	runCmd.Flags().IntVarP(&opt.target, "target", "t", -1, "Set filter definition")
	runCmd.Flags().StringVarP(&opt.logMsg, "write-log", "w", "", "log & comment")
	runCmd.Flags().BoolVarP(&opt.debugMode, "debug", "x", false, "Enable debug output")

}
