package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/exp/constraints"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func abs[T constraints.Signed](a T) T     { return cond(a >= 0, a, -a) }
func max[T constraints.Ordered](a, b T) T { return cond(a >= b, a, b) }
func min[T constraints.Ordered](a, b T) T { return cond(a < b, a, b) }
func cond[T any](t bool, a, b T) T {
	if t == true {
		return a
	} else {
		return b
	}
}

// setEnvVar は環境変数を設定します。
func setEnvVar(key, val string) error {
	err := os.Setenv(key, val)
	if err != nil {
		return err
	}
	return nil
}

// headReader は指定したファイルの先頭行を読み込みます。
func headReader(dir string, fileName string) []string {
	file, err := os.Open(fmt.Sprintf("%s/%s", dir, fileName))
	if err != nil {
		return nil
	}
	defer file.Close()

	var head string
	scanner := bufio.NewScanner(file)
	if scanner.Scan() { // 先頭行のみ読み込み
		head = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return nil
	}
	return strings.Fields(head)
}

// ExecuteWithFileInput はファイルから入力を読み込んでプログラムを実行します。
func ExecuteWithFileInput(filePath string, cmd []string, displayStdout bool, displayStderr bool) (stdout string, stderr string, execErr error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", "", err
	}

	// Set up the command with the program path and pipe the input data to stdin
	var c *exec.Cmd
	if len(cmd) > 1 {
		c = exec.Command(cmd[0], cmd[1:]...)
	} else {
		c = exec.Command(cmd[0])
	}
	c.Stdin = bytes.NewReader(data)
	// Get the output from both stdout and stderr
	var outb, errb bytes.Buffer

	if displayStdout {
		c.Stdout = os.Stdout
		c.Stderr = &errb
	} else if displayStderr {
		c.Stdout = &outb
		c.Stderr = os.Stderr
	} else {
		c.Stdout = &outb
		c.Stderr = &errb
	}
	// Execute the command
	err = c.Run()
	if err != nil {
		return outb.String(), errb.String(), err
	}

	return outb.String(), errb.String(), nil
}

// executeCommand は単一のコマンドを実行し、その出力を返します。
func executeCommand(command []string) ([]byte, error) {
	cmd := exec.Command(command[0], command[1:]...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return o, err
}
func executeCommand2(command string) ([]byte, error) {
	cmdArgs := strings.Fields(command)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return o, err
}

// writeToFile はデータをファイルに書き込みます。
func writeToFile(filename string, data []byte, isAppend bool) error {
	// ファイルオープンモードの選択
	var flag int
	if isAppend {
		flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flag = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	}

	// ファイルオープン
	file, err := os.OpenFile(filename, flag, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// データの書き込み
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// changeDir はカレントディレクトリを変更します。
func changeDir(newDir string) bool {
	previousDirectory, _ = os.Getwd()
	err := os.Chdir(newDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "test:Failed to change directory: %s:%s\n", err, newDir)
		return false
	}

	return true
}

// createDirIfNotExist は指定されたディレクトリが存在しない場合にディレクトリを作成します。
func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755) // 0755 はディレクトリのパーミッションです
	}
	return nil
}

// fileExists は指定したファイルが存在するかどうかを確認します。
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

// downloadFile はURLからファイルをダウンロードします。
func downloadFile(filePath string, url string) error {
	// HTTP GETリクエストを送信
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to download file: " + resp.Status)
	}

	// ファイルを作成
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// レスポンスボディをファイルにコピー
	_, err = io.Copy(out, resp.Body)
	return err
}

// countLines はファイル内の行数を数えます。
func countLines(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0
	}

	return lineCount
}

// ReadLastLine ファイルの最終行を読み込む関数です。
func readLastLine(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var size int64 = 1024 // 読み込むバイト数
	var fi os.FileInfo
	fi, err = file.Stat()
	if err != nil {
		return ""
	}

	// ファイルが小さい場合はファイルサイズを使用
	if fi.Size() < size {
		size = fi.Size()
	}

	bytes := make([]byte, size)
	offset := fi.Size() - size
	_, err = file.ReadAt(bytes, offset)
	if err != nil {
		return ""
	}

	lines := bufio.NewScanner(bufio.NewReader(file))
	var lastLine string
	for lines.Scan() {
		lastLine = lines.Text()
	}
	if err := lines.Err(); err != nil {
		return ""
	}

	return lastLine
}

// stringsToCsv は文字列の配列をCSV形式の文字列に変換します。
func stringsToCsv(s []string) string {
	t := make([]byte, 0)
	for i := 0; i < len(s); i++ {
		t = append(t, []byte(s[i])...)
		t = append(t, ',')
	}
	return string(t)
}

// intsToCsv は整数の配列をCSV形式の文字列に変換します。
func intsToCsv(a []int) string {
	t := make([]byte, 0)
	for i := 0; i < len(a); i++ {
		s := strconv.Itoa(a[i])
		t = append(t, []byte(s)...)
		t = append(t, ',')
	}
	return string(t)
}
func floatsToCsv(f []float64) string {
	t := make([]byte, 0)
	for i := 0; i < len(f); i++ {
		s := fmt.Sprintf("%f", f[i])
		t = append(t, []byte(s)...)
		t = append(t, ',')
	}
	return string(t)
}
func itoa(x int) string {
	return strconv.Itoa(x)
}
func stringTime() string {
	t := time.Now()
	h, m, s := t.Clock()
	_, month, day := t.Date()

	ret := fmt.Sprintf("%02d/%02d %02d:%02d:%02d", month, day, h, m, s)
	return ret
}
func readFileLines(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// バッファサイズを設定
	const maxCapacity = 512 * 102400
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineCount := 0

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return nil
	}

	return lines
}
func renameFile(oldName, newName string) error {
	err := os.Rename(oldName, newName)
	if err != nil {
		return err
	}
	return nil
}

func calcAverage(a []int) (int, int, int) {
	ls := 0.0
	s := 0
	ok := 0
	ng := 0
	for i := 0; i < len(a); i++ {
		if a[i] <= 0 {
			ng++
		} else {
			ok++
			s += a[i]
			ls += math.Log(float64(a[i]))
		}
	}
	if ok == 0 {
		ok = 1
	}
	return int(math.Round(math.Exp(ls / float64(ok)))), s / ok, ng
}

type FileReader struct {
	fileName string
	err      error
	file     *os.File
	scanner  *bufio.Scanner
}

func NewFileReader(fileName string) *FileReader {
	file, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	return &FileReader{
		fileName: fileName,
		err:      nil,
		file:     file,
		scanner:  scanner,
	}
}
func (fr *FileReader) rs() string { fr.scanner.Scan(); return fr.scanner.Text() }
func (fr *FileReader) ri() int {
	fr.scanner.Scan()
	i, e := strconv.Atoi(fr.scanner.Text())
	if e != nil {
		panic(e)
	}
	return i
}
func (fr *FileReader) rf() float64 {
	f, e := strconv.ParseFloat(fr.rs(), 64)
	if e != nil {
		panic(e)
	}
	return f
}
func (fr *FileReader) close() {
	fr.file.Close()
}
func insertLine(filename string, lineNumber int, newLine string) error {
	// ファイルを開く
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("ファイルを開く際にエラーが発生しました: %v", err)
	}
	defer file.Close()

	// ファイル全体を読み込む
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ファイルを読み込む際にエラーが発生しました: %v", err)
	}

	// 新しい行を指定した位置に挿入
	if lineNumber > len(lines) {
		lineNumber = len(lines) + 1
	}
	lines = append(lines[:lineNumber-1], append([]string{newLine}, lines[lineNumber-1:]...)...)

	// ファイルに書き込み
	file, err = os.Create(filename)
	if err != nil {
		return fmt.Errorf("ファイルを作成する際にエラーが発生しました: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("ファイルに書き込む際にエラーが発生しました: %v", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("ファイルをフラッシュする際にエラーが発生しました: %v", err)
	}

	return nil
}

func dbg(file string, s ...interface{}) {
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		msg := fmt.Sprintf("%s:%d:%s\n", funcName, line, fmt.Sprint(s))
		writeToFile(file, []byte(msg), true)
	}
}
