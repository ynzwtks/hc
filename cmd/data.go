package cmd

import (
	"github.com/schollz/progressbar/v3"
	"time"
)

var version string

const (
	padding  = 2
	maxWidth = 80
	INF      = 1 << 60
)

const ResultCsv = "result.csv"
const ContestToml = "contest.toml"
const HistoryCsv = "history.csv"
const RunCsv = "run.csv"
const InputCsv = "input.csv"
const MaxHistoryRefSize = 10000

var confPath string

type Config struct {
	Common    Common             `toml:"common"`
	TestSets  map[string]TestSet `toml:"-"`
	Standings Standings          `toml:"standings"`
	Cloud     Cloud              `toml:"cloud"`
	Env       Env                `toml:"env"`
}

type Common struct {
	ContestName   string `toml:"ContestName"`
	TargetProgram string `toml:"TargetProgram"`
	JudgeProgram  string `toml:"JudgeProgram"`
	GenProgram    string `toml:"GenProgram"`
	BaseDir       string `toml:"BaseDir"`
	BuildCmd      string `toml:"BuildCmd"`
	InputFields   string `toml:"InputFields"`
	IsInteractive bool   `toml:"IsInteractive"`
	Workers       int    `toml:"Workers"`
	DefaultSet    string `toml:"DefaultSet"`
	IsRankMin     bool   `toml:"IsRankMin"`
	ScoreLine     string `toml:"ScoreLine"`
}

type TestSet struct {
	SetName      string   `toml:"SetName"`
	TestDataPath string   `toml:"TestDataPath"`
	TestDataNum  int      `toml:"TestDataNum"`
	ExFields     string   `toml:"ExFields"`
	Seeds        []string `toml:"-"`
	IsSystemTest bool     `toml:"IsSystemTest"`
}

type Standings struct {
	Enable        bool   `toml:"Enable"`
	IndexHtmlURL  string `toml:"IndexHtmlURL"`
	VisualizerURL string `toml:"VisualizerURL"`
	RelEval       int    `toml:"RelEval"`
}

type Cloud struct {
	SetName          string   `toml:"SetName"`
	BuildCmd         string   `toml:"BuildCmd"`
	TargetProgramX64 string   `toml:"TargetProgramX64"`
	ImageURL         string   `toml:"ImageURL"`
	ServiceAccount   string   `toml:"ServiceAccount"`
	BucketName       string   `toml:"BucketName"`
	JobName          string   `toml:"JobName"`
	JobCounts        int      `toml:"JobCounts"`
	JobRegions       []string `toml:"JobRegions"`
	JobBase          []int    `toml:"JobBase"`
	JobTasks         []int    `toml:"JobTasks"`
	JobStep          []int    `toml:"JobStep"`
}
type Env struct {
	Keys   []string `toml:"Keys"`
	Values []string `toml:"Values"`
}
type HeaderInfo struct {
	Header     []string
	HeaderData [][]string
}
type Options struct {
	setName       string
	filter        string
	loop          int
	target        int
	logMsg        string
	asc           bool
	order         string
	linesLimit    int
	quietMode     bool
	testSeedBegin int
	testCount     int
}
type SetupOptions struct {
	setName       string
	testSeedBegin int
	testCount     int
	contestName   string
}
type RuntimeInfo struct {
	caption            []string
	testID             []int
	score              []pair
	failedTask         []string
	lastDist           []int
	bestDist           []int
	incLast            []scoreElem
	incBest            []scoreElem
	decLast            []scoreElem
	decBest            []scoreElem
	scoreSum           int
	scoreLogSum        float64
	okCnt              int
	ngCnt              int
	ng                 []int
	vsBest             int
	vsLast             int
	bar                *progressbar.ProgressBar
	executed           int
	executingCase      []string
	enableLog          bool
	enableLogStandings bool
	lastDisplayTime    time.Time
}
type Logs struct {
	logRootDir      string
	logDir          string
	last            []int
	best            []int
	vals            [][]int
	idxes           []int
	isBlank         bool
	times           []string
	comments        []string
	best2           []int
	vals2           [][]int
	idxes2          []int
	isBlank2        bool
	lastDisplayTime time.Time
}

var conf Config
var cmn Common
var logs Logs
var set TestSet
var opt Options
var setupOpt SetupOptions
var ri RuntimeInfo
var sd Standings
var jobs Cloud
var hi HeaderInfo
var env Env
var previousDirectory string

type pair struct{ a, b int }

type scoreElem struct {
	id       string
	ratio    float64
	newScore int
	oldScore int
}

var configTemplate = `
[common]
ContestName = "ContestName"
BaseDir  = "."
BuildCmd = ""
IsInteractive = true
TargetProgram = ""
JudgeProgram = ""
InputFields = ""
Workers = 5
DefaultSet = ""
IsRankMin = true
ScoreLine = "Score ="
[standings]
Enable = true
IndexHtmlURL = "https://img.atcoder.jp/ahc_standings/index.html"
VisualizerURL =""
RelEval = 100000000
[cloud]
SetName = ""
BuildCmd = ""
TargetProgramX64 = ""
ImageURL = ""
BucketName = ""
ServiceAccount=""
JobName = ""
JobCounts = 0
JobRegions =[]
JobBase = []
JobTasks = []
JobStep = []
`
