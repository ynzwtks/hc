package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure settings",
	Long:  `Configure settings`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "configure settings",
	Long:  `configure settings`,
	Run: func(cmd *cobra.Command, args []string) {
		readConf()
		prompts := []*survey.Question{
			{
				Name: "ContestName",
				Prompt: &survey.Input{
					Message: "Enter the contest display name(Required):",
					Default: conf.Common.ContestName,
				},
				Validate: survey.Required,
			},
			{
				Name: "BaseDir",
				Prompt: &survey.Input{
					Message: "Enter the base directory(Required):",
					Default: conf.Common.BaseDir,
				},
				Validate: survey.Required,
			},
			{
				Name: "BuildCmd",
				Prompt: &survey.Input{
					Message: "Enter the build command(Optional):",
					Default: conf.Common.BuildCmd,
				},
			},
			{
				Name: "TargetProgram",
				Prompt: &survey.Input{
					Message: "Enter the path to the program to run for the local test(Required):",
					Default: conf.Common.TargetProgram,
				},
				Validate: survey.Required,
			},
			{
				Name: "JudgeProgram",
				Prompt: &survey.Input{
					Message: "Enter the judge program(Tester or Vis) path(Required):",
					Default: conf.Common.JudgeProgram,
				},
				Validate: survey.Required,
			},
			{
				Name: "GenProgram",
				Prompt: &survey.Input{
					Message: "Enter the Input Generator(gen) path(Required):",
					Default: conf.Common.GenProgram,
				},
				Validate: survey.Required,
			},

			{
				Name: "InputFields",
				Prompt: &survey.Input{
					Message: "Enter the input header's fields(Optional):",
					Default: conf.Common.InputFields,
				},
			},
			{
				Name: "IsInteractive",
				Prompt: &survey.Confirm{
					Message: "Is this contest interactive type?",
					Default: conf.Common.IsInteractive,
				},
			},
			{
				Name: "Workers",
				Prompt: &survey.Input{
					Message: "Enter the number of workers(Required):",
					Default: fmt.Sprintf("%d", conf.Common.Workers),
				},
				Validate: survey.Required,
			},
		}
		err := survey.Ask(prompts, &conf.Common)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		UpdateConfig()
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "list test sets",
	Long:  `list test sets`,
	Run: func(cmd *cobra.Command, args []string) {
		readConf()
		defaultSet := conf.Common.DefaultSet
		titles := []string{"Set Name", "Size", "Path"}
		data := [][]string{}
		for _, v := range conf.TestSets {
			t := []string{v.SetName, itoa(v.TestDataNum), v.TestDataPath}
			data = append(data, t)
		}
		table := createTestSetTable(titles, data, defaultSet)
		println(table)
	},
}

func createTestSetTable(title []string, data [][]string, defaultName string) string {
	// 各列の幅を設定
	columnWidths := []int{20, 10, 20}

	// タイトルのセルスタイル定義
	titleStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		Align(lipgloss.Left).
		Bold(true) // タイトルのテキスト色

	// タイトル行の作成
	titleRowCells := []string{}
	for i, t := range title {
		titleText := titleStyle.Width(columnWidths[i]).Render(t)
		titleRowCells = append(titleRowCells, titleText)
	}
	titleRow := lipgloss.JoinHorizontal(lipgloss.Top, titleRowCells...)

	// データのセルスタイル定義
	dataCellStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		Align(lipgloss.Left)

	// データ行の作成
	dataRows := ""
	for _, row := range data {
		dataRowCells := []string{}
		if row[0] == defaultName {
			dataRowCells = append(dataRowCells, lipgloss.NewStyle().Render("✓"))
		} else {
			dataRowCells = append(dataRowCells, lipgloss.NewStyle().Render(" "))
		}
		for i, col := range row {

			dataRowCells = append(dataRowCells, dataCellStyle.Width(columnWidths[i]).Foreground(lipgloss.Color("white")).Render(col))

		}
		dataRow := lipgloss.JoinHorizontal(lipgloss.Top, dataRowCells...)
		dataRows += dataRow + "\n"
	}

	// 完成した表を返す
	return titleRow + "\n" + dataRows
}

var configSwitch = &cobra.Command{
	Use:   "switch",
	Short: "switch the default test set",
	Long:  `switch the default test set`,
	Run: func(cmd *cobra.Command, args []string) {
		readConf()
		conf.Common.DefaultSet = setupOpt.setName
		UpdateConfig()
		fmt.Println("default test set :", conf.Common.DefaultSet)
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a test set",
	Long:  `add a test set`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
var configRemove = &cobra.Command{
	Use:   "remove",
	Short: "remove a test set",
	Long:  `remove a test set`,
	Run: func(cmd *cobra.Command, args []string) {
		removeTestSet(setupOpt.setName)
	},
}
var configAddTest = &cobra.Command{
	Use:   "test",
	Short: "generate a test set",
	Long:  `generate a test set`,
	Run: func(cmd *cobra.Command, args []string) {
		readConf()
		genTestSet(setupOpt.testSeedBegin, setupOpt.testCount, setupOpt.setName)
	},
}
var configAddSystemTest = &cobra.Command{
	Use:   "system",
	Short: "add a system test",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		setupSystemTest(setupOpt.contestName)
	},
}

func setupSystemTest(contestName string) {
	readConf()
	changeDir(cmn.BaseDir)

	seedsURL := fmt.Sprintf("https://img.atcoder.jp/%s/seeds.txt", strings.ToLower(contestName))
	inputCsvURL := fmt.Sprintf("https://img.atcoder.jp/ahc_standings/%s/input.csv", strings.ToLower(contestName))
	resultCsvURL := fmt.Sprintf("https://img.atcoder.jp/ahc_standings/%s/result.csv", strings.ToLower(contestName))

	logsPath := fmt.Sprintf("logs/system")
	testPath := fmt.Sprintf("test/system")
	createDirIfNotExist(logsPath)
	createDirIfNotExist(testPath)
	seeds := fmt.Sprintf("%s/seeds.txt", testPath)

	inputCsv := fmt.Sprintf("%s/input.csv", logsPath)
	resultCsv := fmt.Sprintf("%s/result.csv", logsPath)
	err := downloadFile(seeds, seedsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to download seed file:%s\n", seedsURL)
		os.Exit(1)
	}

	cmd := []string{cmn.GenProgram, seeds, "-d", testPath}
	_, err = executeCommand(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create seeds file: %s\n", cmd)
		os.Exit(1)
	}
	err = downloadFile(inputCsv, inputCsvURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Not Found :%s\n", inputCsvURL)
	}

	err = downloadFile(resultCsv, resultCsvURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Not found : %s\n", resultCsvURL)
	}
	t := TestSet{}
	t.TestDataNum = countLines(seeds)
	t.SetName = "system"
	t.TestDataPath = testPath
	t.IsSystemTest = true
	conf.TestSets["system"] = t
	cmn.DefaultSet = t.SetName
	UpdateConfig()

	fmt.Println("System test definition added.")
}
func genTestSet(begin, cnt int, setName string) {
	readConf()
	changeDir(cmn.BaseDir)
	_, ok := conf.TestSets[setName]
	if ok {
		fmt.Println("Test set already exists.")
		return
	}
	testPath := fmt.Sprintf("test/%s", setName)
	createDirIfNotExist(testPath)
	seeds := fmt.Sprintf("%s/seeds.txt", testPath)

	dat := make([]byte, 0)
	for i := begin; i < begin+cnt; i++ {
		dat = append(dat, []byte(fmt.Sprintf("%d\n", i))...)
	}
	writeToFile(seeds, dat, false)
	cmd := []string{cmn.GenProgram, seeds, "-d", testPath}
	executeCommand(cmd)

	t := TestSet{}
	t.TestDataNum = cnt
	t.SetName = setName
	t.IsSystemTest = false
	t.TestDataPath = testPath
	conf.TestSets[setName] = t
	conf.Common.DefaultSet = setName
	UpdateConfig()
	fmt.Println("Test definition added.")
}
func removeTestSet(setName string) {
	if len(setName) == 0 {
		return
	}
	readConf()
	changeDir(cmn.BaseDir)
	testPath := fmt.Sprintf("test/%s", setName)
	logPath := fmt.Sprintf("log/%s", setName)
	err := os.RemoveAll(testPath)
	if err != nil {
		return
	}
	err = os.RemoveAll(logPath)
	if err != nil {
		return
	}
	_, ok := conf.TestSets[setName]
	if ok {
		delete(conf.TestSets, setName)
	}
	if setName == conf.Common.DefaultSet {
		setName = ""
		for k := range conf.TestSets {
			setName = k
			break
		}
		conf.Common.DefaultSet = setName
	}
	UpdateConfig()
}

var configNewCmd = &cobra.Command{
	Use:   "new",
	Short: "create the contest configuration template",
	Long:  "create the contest configuration template",
	Run: func(cmd *cobra.Command, args []string) {
		if fileExists(ContestToml) {
			renameFile(ContestToml, ContestToml+".1")
		}
		err := writeToFile(ContestToml, []byte(configTemplate), false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create the configuration file (template:./%s)\n", ContestToml)
			os.Exit(1)
		}
		fmt.Printf("Successfully created the configuration file (template: ./%s)\n", ContestToml)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetupCmd)
	configCmd.AddCommand(configNewCmd)
	configCmd.AddCommand(configAddCmd)

	configCmd.AddCommand(configRemove)
	configRemove.Flags().StringVarP(&setupOpt.setName, "setName", "s", "", "Set the name of the configuration")
	configRemove.MarkFlagRequired("setName")

	configCmd.AddCommand(configSwitch)
	configSwitch.Flags().StringVarP(&setupOpt.setName, "setName", "s", "", "Set the name of the configuration")
	configSwitch.MarkFlagRequired("setName")

	configCmd.AddCommand(configListCmd)
	configAddCmd.AddCommand(configAddTest)
	configAddTest.Flags().StringVarP(&setupOpt.setName, "setName", "s", "", "Set the name of the configuration")
	configAddTest.Flags().IntVarP(&setupOpt.testCount, "count", "c", 0, "Set the count")
	configAddTest.Flags().IntVarP(&setupOpt.testSeedBegin, "begin", "b", 0, "Set the beginning value (default is 0)")

	configAddTest.MarkFlagRequired("setName")
	configAddTest.MarkFlagRequired("count")

	configAddCmd.AddCommand(configAddSystemTest)
	configAddSystemTest.Flags().StringVarP(&setupOpt.contestName, "contestName", "n", "", "Set the name of the contest")
	configAddSystemTest.MarkFlagRequired("contestName")
}
