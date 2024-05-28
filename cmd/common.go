package cmd

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func commonInit() {
	readConf()
	changeDir(cmn.BaseDir)
	readParameter()
	logsInit()
	loadLogs()
}

func loadLogs() {
	loadHistoryCsv()
	loadResultCsv()

}
func loadResultCsv() {
	logs.best2 = make([]int, set.TestDataNum)
	for i := 0; i < set.TestDataNum; i++ {
		logs.best2[i] = -1
	}
	csvFile := fmt.Sprintf("%s/%s", logs.logDir, ResultCsv)

	lc := readFileLines(csvFile)
	if len(lc) <= 1 {
		logs.isBlank2 = true
		return
	}

	lc = lc[1:]
	if len(lc) > MaxHistoryRefSize {
		lc = lc[max(0, len(lc)-MaxHistoryRefSize):]
	}
	results := make([][]int, len(lc))
	for i := 0; i < len(lc); i++ {
		ls := strings.Split(lc[i], ",")
		t := ls[1:]
		idx := i
		logs.idxes2 = append(logs.idxes2, idx)
		t2 := make([]int, set.TestDataNum)
		for j := 0; j < set.TestDataNum; j++ {
			v, _ := strconv.Atoi(t[j])
			t2[j] = v
			if logs.best2[j] <= 0 && v > 0 {
				logs.best2[j] = v
				continue
			}
			if cmn.IsRankMin {
				if logs.best2[j] > v && v > 0 {
					logs.best2[j] = v
				}
			} else {
				if logs.best2[j] < v && v > 0 {
					logs.best2[j] = v
				}
			}
		}
		results[i] = t2
	}
	logs.vals2 = results
}

func loadHistoryCsv() {
	logs.best = make([]int, set.TestDataNum)
	logs.last = make([]int, set.TestDataNum)
	for i := 0; i < set.TestDataNum; i++ {
		logs.best[i] = -1
		logs.last[i] = -1
	}
	historyCsv := fmt.Sprintf("%s/history.csv", logs.logDir)
	lc := readFileLines(historyCsv)
	if len(lc) < 1 {
		logs.isBlank = true
		return
	}
	if len(lc) > MaxHistoryRefSize {
		lc = lc[max(0, len(lc)-MaxHistoryRefSize):]
	}
	results := make([][]int, len(lc))
	for i := 0; i < len(lc); i++ {
		ls := strings.Split(lc[i], ",")
		t := ls[3:]
		logs.times = append(logs.times, ls[0])
		logs.comments = append(logs.comments, ls[2])
		idx, _ := strconv.Atoi(ls[1])
		logs.idxes = append(logs.idxes, idx)
		t2 := make([]int, set.TestDataNum)
		for j := 0; j < set.TestDataNum; j++ {
			v, _ := strconv.Atoi(t[j])
			t2[j] = v
			if i == len(lc)-1 {
				logs.last[j] = v
			}
			if logs.best[j] <= 0 && v > 0 {
				logs.best[j] = v
				continue
			}
			if cmn.IsRankMin {
				if logs.best[j] > v && v > 0 {
					logs.best[j] = v
				}
			} else {
				if logs.best[j] < v && v > 0 {
					logs.best[j] = v
				}
			}
		}
		results[i] = t2
	}
	logs.vals = results

}

func readConf() {
	viper.SetConfigName("contest") // 設定ファイルの名前（拡張子を除く）
	viper.SetConfigType("toml")    // 設定ファイルの形式
	viper.AddConfigPath(".")       // 設定ファイルのパス（ここではカレントディレクトリ）

	var err error

	confPath = "./contest.toml"
	if err = viper.ReadInConfig(); err != nil {
		envConfigPath := os.Getenv("CONTEST_CONFIG_FILE")
		if envConfigPath != "" {
			// 環境変数が設定されている場合、そのパスを利用
			confPath = envConfigPath
			viper.SetConfigFile(envConfigPath)
			if err = viper.ReadInConfig(); err != nil {
				log.Fatalf("Error reading config file, %s", err)
			}
		}
	}
	settings := viper.AllSettings()
	conf.TestSets = make(map[string]TestSet)
	for key, value := range settings {
		switch key {
		case "common":
			mapstructure.Decode(value, &conf.Common)
			cmn = conf.Common
		case "cloud":
			mapstructure.Decode(value, &conf.Cloud)
			jobs = conf.Cloud
		case "standings":
			mapstructure.Decode(value, &conf.Standings)
			sd = conf.Standings
		case "env":
			mapstructure.Decode(value, &conf.Env)
			env = conf.Env
			if len(env.Keys) > 0 && len(env.Keys) == len(env.Values) {
				for i := 0; i < len(env.Keys); i++ {
					setEnvVar(env.Keys[i], env.Values[i])
				}
			}
		default:
			var s TestSet
			mapstructure.Decode(value, &s)
			conf.TestSets[key] = s
		}
	}

	if opt.setName == "default" {
		opt.setName = cmn.DefaultSet
	}
	var ret bool
	set, ret = conf.TestSets[opt.setName]
	if ret == false {
		for k, v := range conf.TestSets {
			opt.setName = k
			set = v
		}
	}

}

func filterEvaluate(s []string) bool {
	if len(opt.filter) == 0 {
		return true
	}
	parameters := make(map[string]interface{}, 8)
	for i := 0; i < len(s); i++ {
		v, err := strconv.ParseFloat(s[i], 64)
		if err == nil {

			parameters[hi.Header[i]] = v
		}
	}
	filter, _ := govaluate.NewEvaluableExpression(opt.filter)
	result, _ := filter.Evaluate(parameters)
	if result == false {
		return false
	} else {
		return true
	}
}

func readParameter() {
	//hi.HeaderData = make([][]string, set.TestDataNum)

	set.Seeds = readFileLines(fmt.Sprintf("%s/%s", set.TestDataPath, "seeds.txt"))

	ri.caption = append(ri.caption, "")
	fs := strings.Fields(cmn.InputFields)

	exHeader, exDat := readExParam(set.TestDataPath, set.TestDataNum, set.ExFields)
	if len(exHeader) != 0 && len(exDat) == set.TestDataNum {
		fs = append(fs, exHeader...)
	}
	hi.Header = fs
	for i := 0; i < set.TestDataNum; i++ {
		fileName := fmt.Sprintf("%04d.txt", i)
		ret := headReader(set.TestDataPath, fileName)
		// Todo exDatのエラー処理を追加する
		if len(exHeader) != 0 && len(exDat) == set.TestDataNum {
			ret = append(ret, exDat[i]...)
		}
		hi.HeaderData = append(hi.HeaderData, ret)
	}
	if opt.target != -1 {
		ri.testID = append(ri.testID, int(opt.target))
	}
	for i := 0; i < set.TestDataNum; i++ {
		if filterEvaluate(hi.HeaderData[i]) == false {
			continue
		}
		ri.testID = append(ri.testID, i)
	}

	//panic("hoge")
}

func logsInit() {
	logs.logRootDir = fmt.Sprintf("%s/%s", cmn.BaseDir, "logs")
	createDirIfNotExist(logs.logDir)
	logs.logDir = fmt.Sprintf("%s/%s", logs.logRootDir, set.SetName)
	createDirIfNotExist(logs.logDir)

	if sd.Enable {
		seedsFile := fmt.Sprintf("%s/seeds.txt", set.TestDataPath)
		set.Seeds = readFileLines(seedsFile)

		indexHtml := fmt.Sprintf("%s/index.html", logs.logRootDir)
		ret := fileExists(indexHtml)
		if !ret {
			downloadFile(indexHtml, sd.IndexHtmlURL)
		}
		inputCsv := fmt.Sprintf("%s/input.csv", logs.logDir)
		ret = fileExists(inputCsv)
		if !ret {
			header := fmt.Sprintf("file,seed,%s\n", stringsToCsv(hi.Header))
			writeToFile(inputCsv, []byte(header), false)
			var line string
			for i := 0; i < set.TestDataNum; i++ {
				if len(set.Seeds) == 0 {
					line = fmt.Sprintf("%04d,%s,%s\n", i, i, stringsToCsv(hi.HeaderData[i]))
				} else {
					line = fmt.Sprintf("%04d,%s,%s\n", i, set.Seeds[i], stringsToCsv(hi.HeaderData[i]))
				}

				writeToFile(inputCsv, []byte(line), true)
			}
		}

		resultCsv := fmt.Sprintf("%s/result.csv", logs.logDir)
		ret = fileExists(resultCsv)
		if !ret {
			header := fmt.Sprintf("rank_min,%d,%s\n", sd.RelEval, sd.VisualizerURL)
			writeToFile(resultCsv, []byte(header), false)
		}
	}
}

func printLineBack(n int) {
	lineBack := "\u001B[A\u001B[K"
	for i := 0; i < n; i++ {
		fmt.Print(lineBack)
	}
}
func printLine(n int) {
	for i := 0; i < n; i++ {
		fmt.Println("")
	}
}
func printLargeScore(n int) {
	sort.Slice(ri.score, func(i, j int) bool {
		return ri.score[i].b > ri.score[j].b
	})
	for i := 0; i < min(n, len(ri.score)); i++ {
		fmt.Printf("%04d %d \n", ri.score[i].a, ri.score[i].b)
	}
	fmt.Println()
}
func printTitle() {
	dl := len(cmn.ContestName) + len(opt.setName) + 2
	var style = lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("7")).
		Align(lipgloss.Center).
		Width(dl + 4)
	fmt.Printf("%s", style.Render(fmt.Sprintf("%s(%s)", cmn.ContestName, opt.setName)))
	printLine(16)
}

func readExParam(testDataPath string, testDataNum int, fields string) ([]string, [][]string) {
	fs := strings.Fields(fields)
	if len(fs) == 0 {
		return nil, nil
	}
	fr := NewFileReader(fmt.Sprintf("%s/ex.dat", testDataPath))
	if fr == nil {
		return nil, nil
	}
	defer fr.close()
	ret := make([][]string, testDataNum)
	for i := 0; i < len(ret); i++ {
		ret[i] = make([]string, len(fields))
		for j := 0; j < len(fields); j++ {
			ret[i][j] = fr.rs()
		}
	}
	return fs, ret
}
func buildCmd(cmd string) bool {
	if len(cmd) == 0 {
		return true
	}
	_, err := executeCommand2(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute command '%s': %s\n", cmd, err)
		os.Exit(1)
	}
	return true
}
func UpdateConfig() {
	f, err := os.Create(confPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(conf); err != nil {
		log.Fatal(err)
	}
	for setName, testData := range conf.TestSets {
		fmt.Fprintf(f, "[%s]\n", setName)
		if err := toml.NewEncoder(f).Encode(testData); err != nil {
			log.Fatal(err)
		}
	}
}
func calcRank(sc, no int) int {
	rank := 1
	if len(logs.vals2) != 0 && set.IsSystemTest {

		for i := 0; i < len(logs.vals2); i++ {
			if logs.vals2[i][no] <= 0 {
				continue
			}
			if cmn.IsRankMin {
				if logs.vals2[i][no] < sc {
					rank++
				}
			} else {
				if logs.vals2[i][no] > sc {
					rank++
				}
			}
		}
		return rank
	} else {
		for i := 0; i < len(logs.vals); i++ {
			if logs.vals[i][no] <= 0 {
				continue
			}
			if cmn.IsRankMin {
				if logs.vals[i][no] < sc {
					rank++
				}
			} else {
				if logs.vals[i][no] > sc {
					rank++
				}
			}
		}
		return rank
	}
}
