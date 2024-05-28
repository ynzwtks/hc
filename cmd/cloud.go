package cmd

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/iterator"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// cloudCmd represents the cloud command
var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Execute cloud run jobs",
	Long:  `Execute cloud run jobs`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var jobsRunCmd = &cobra.Command{
	Use:   "run",
	Short: "execute cloud run jobs",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if opt.logMsg == "" {
			return fmt.Errorf("-w/--write-log is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("initializing")

		// Init処理
		readConf()
		err := viper.UnmarshalKey(jobs.SetName, &set)
		if err != nil {
			log.Fatalf("Unable to decode into struct, %s", err)
		}
		changeDir(cmn.BaseDir)
		readParameter()
		runtimeInit()
		logsInit()
		loadLogs()

		// x86-64向けのコンパイルしてオブジェクトストレージにアップロードする
		if !buildCmd(jobs.BuildCmd) {
			return
		}
		err = uploadFile(jobs.BucketName, fmt.Sprintf("%s/input/", jobs.JobName), jobs.TargetProgramX64)
		if err != nil {
			log.Fatalf("Unable to upload the target progam, %s", err)
		}
		// オブジェクトストレージの前回の結果ファイルを削除する
		deleteObjectsInFolder()
		runJobs()
		resultReader()
		printLog()
	},
}
var jobsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete cloud run jobs definition",
	Long:  `delete cloud run jobs definition`,
	Run: func(cmd *cobra.Command, args []string) {
		readConf()
		deleteJobs()
	},
}
var jobsListCmd = &cobra.Command{
	Use:   "list",
	Short: "list cloud run jobs definitions",
	Long:  `list cloud run jobs definitions`,
	Run: func(cmd *cobra.Command, args []string) {
		readConf()
		listJobs()
	},
}
var jobsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create cloud run jobs definition",
	Long:  `create cloud run jobs definition`,
	Run: func(cmd *cobra.Command, args []string) {
		readConf()
		createJobs()
	},
}

func deleteJobs() {
	readConf()
	for i := 0; i < jobs.JobCounts; i++ {
		cmd := exec.Command("gcloud", "run", "jobs", "delete", jobs.JobName, "--quiet", "--region", jobs.JobRegions[i])
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("エラーが発生しました(delete): %s\n", err)
		}
		fmt.Println(string(output))
	}
}

func listJobs() {
	cmd := exec.Command("gcloud", "run", "jobs", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("エラーが発生しました(delete): %s\n", err)
	}
	fmt.Println(string(output))
}

func createJobs() {
	for i := 0; i < len(jobs.JobRegions); i++ {

		var1 := fmt.Sprintf("BUCKET_NAME=%s", jobs.BucketName)
		var2 := fmt.Sprintf("JOB_NAME=%s", jobs.JobName)
		var3 := fmt.Sprintf("BASE=%d", jobs.JobBase[i])
		var4 := fmt.Sprintf("STEP=%d", jobs.JobStep[i])

		cmd := exec.Command("gcloud", "run", "jobs", "create", "ahc031",
			"--quiet",
			"--region", jobs.JobRegions[i],
			"--task-timeout", "10m",
			"--max-retries", "0",
			"--cpu", "2",
			"--memory", "2G",
			"--set-env-vars", var1,
			"--set-env-vars", var2,
			"--tasks", "50",
			"--set-env-vars", var3,
			"--set-env-vars", var4,
			"--service-account", jobs.ServiceAccount,
			"--image", jobs.ImageURL,
		)
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("エラーが発生しました(create): %s\n", err)
		}
		fmt.Println(string(output))
	}
}

func runJobs() {

	numWorkers := jobs.JobCounts
	ri.executingCase = make([]string, numWorkers)

	// タスク用のチャネル
	tasks := make(chan int, numWorkers)

	// WaitGroupの初期化
	var wg sync.WaitGroup

	// ワーカーの起動
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go executeJob(&wg, tasks)
	}

	for i := 0; i < jobs.JobCounts; i++ {
		tasks <- i
	}
	close(tasks)
	wg.Wait()

}

func executeJob(wg *sync.WaitGroup, tasks <-chan int) {
	defer wg.Done()
	for task := range tasks {
		log.Printf("started job @%s Base=%-4d Task=%d Step=%d", jobs.JobRegions[task], jobs.JobBase[task], jobs.JobTasks[task], jobs.JobStep[task])
		cmd := exec.Command("gcloud", "run", "jobs", "execute", jobs.JobName, "--region", jobs.JobRegions[task], "--wait")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("エラーが発生しました(delete): %s\n", err)
		}
		log.Printf("finished job @%s", jobs.JobRegions[task])
		_ = output
		//fmt.Println(string(output))
	}
}
func deleteObjectsInFolder() error {
	bucketName := jobs.BucketName
	folderPrefix := fmt.Sprintf("%s/out", jobs.JobName)
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	var wg sync.WaitGroup
	bucket := client.Bucket(bucketName)
	query := &storage.Query{Prefix: folderPrefix}

	it := bucket.Objects(ctx, query)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(objectName string) {
			defer wg.Done()
			if err := bucket.Object(objectName).Delete(ctx); err != nil {
				log.Printf("Failed to delete object: %s, error: %v\n", objectName, err)
			}
		}(objAttrs.Name)
	}
	wg.Wait()
	log.Printf("Deleted previous results in folder: %s\n", folderPrefix)
	return nil
}

func resultReader() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close()
	mu := sync.Mutex{} // マップへのアクセスを同期するためのミューテックス
	bucketName := jobs.BucketName
	prefix := fmt.Sprintf("%s/out/", jobs.JobName)
	var wg sync.WaitGroup
	it := client.Bucket(bucketName).Objects(ctx, &storage.Query{Prefix: prefix})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Error listing objects: %v\n", err)
			return
		}
		wg.Add(1)
		go func(objectName string) {
			defer wg.Done()
			rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
			if err != nil {
				fmt.Printf("Error opening object: %v\n", err)
				return
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				fmt.Printf("Error reading object: %v\n", err)
				return
			}

			// テスト結果を処理
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				parts := strings.Split(line, " ")
				if len(parts) == 2 {
					idx, _ := strconv.Atoi(parts[0])
					sc, _ := strconv.Atoi(parts[1])
					mu.Lock()
					ri.scoreSum += sc
					ri.score[idx].a = idx
					ri.score[idx].b = max(ri.score[idx].b, sc)
					if sc == 0 {
						ri.ngCnt++
					} else {
						ri.okCnt++
					}
					mu.Unlock()
				}
			}
		}(attrs.Name)
	}
	wg.Wait()
	log.Printf("collected job results")
}
func uploadFile(bucketName, objectName, srcFilePath string) error {
	// コンテキストとクライアントを作成
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// バケットとオブジェクトを指定
	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)

	// ファイルをアップロードするためのWriterを作成
	wc := object.NewWriter(ctx)
	defer wc.Close()

	// アップロードするファイルを開く
	f, err := os.Open(srcFilePath)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	// ファイルの内容をWriterにコピー
	if _, err := io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	log.Printf("File %q has been uploaded to %q.", srcFilePath, objectName)
	return nil
}
func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.AddCommand(jobsRunCmd)
	jobsRunCmd.Flags().StringVarP(&opt.logMsg, "write-log", "w", "", "log & comment")
	jobsCmd.AddCommand(jobsDeleteCmd)
	jobsCmd.AddCommand(jobsListCmd)
	jobsCmd.AddCommand(jobsCreateCmd)
}
