package judge

import (
	"fmt"
	"github.com/gogather/com"
	"github.com/gojudge/judger/core"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var BuildStartTime int64
var BuildProcessHaveExit bool

type Compile struct {
	system        string
	buildPath     string
	userBuildPath string
	itemBuildPath string
	codeFilePath  string
	currentPath   string

	compiler_c string
	postfix_c  string

	compiler_cpp string
	postfix_cpp  string

	buildOverTime bool
}

func (this *Compile) NewCompile() {
	this.buildOverTime = false
	this.system = runtime.GOOS
	this.postfix_c = "c"
	this.postfix_cpp = "cpp"
	this.currentPath, _ = os.Getwd()

	this.buildPath = filepath.Join(this.currentPath, core.C.Get(runtime.GOOS, "buildpath"))
	this.compiler_c = filepath.Join(this.currentPath, core.C.Get(runtime.GOOS, "compiler_c"))

	log.Println("[current path]", this.currentPath)
	log.Println("[build path]", this.buildPath)
	log.Println("[compiler path]", this.compiler_c)
}

func (this *Compile) Run(code string, language string, id int, sid string) (string, error) {

	err := this.createDirs(id, sid)
	if err != nil {
		log.Println(err)
		return "", err
	} else {
		err = this.writeCode(code, id, language)
		if err != nil {
			log.Println(err)
			return "", err
		}
	}

	return this.itemBuildPath, this.gcc(id)

}

// 创建编译环境的目录结构
func (this *Compile) createDirs(id int, sid string) error {
	var err error
	err = nil
	this.userBuildPath = filepath.Join(this.buildPath, sid)
	if !com.PathExist(this.userBuildPath) {
		err = com.Mkdir(this.userBuildPath)
	}
	this.itemBuildPath = filepath.Join(this.userBuildPath, fmt.Sprintf("%d", id))
	if !com.PathExist(this.itemBuildPath) {
		err = com.Mkdir(this.itemBuildPath)
	}
	return err
}

// 代码写入文件
func (this *Compile) writeCode(code string, id int, language string) error {
	lang := ""
	if language == "C" {
		lang = "c"
	}
	this.codeFilePath = filepath.Join(this.itemBuildPath, fmt.Sprintf("%d.%s", id, lang))
	return com.WriteFile(this.codeFilePath, code)
}

// call gcc compiler in other os
func (this *Compile) gcc(id int) error {
	os.Chdir(this.itemBuildPath)

	var cmd *exec.Cmd

	log.Println("cmd", "/K",
		this.compiler_c,
		this.codeFilePath,
		this.itemBuildPath,
	)

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/K",
			this.compiler_c,
			this.codeFilePath,
			this.itemBuildPath,
		)
	} else {
		cmd = exec.Command("sh",
			this.compiler_c,
			this.codeFilePath,
			this.itemBuildPath,
		)
	}

	err := cmd.Start()
	if err != nil {
		log.Println("Start Failed")
		log.Println(err)
	}

	stn := time.Now()
	BuildStartTime = stn.UnixNano()
	go checkTimer(cmd, this, id)
	BuildProcessHaveExit = false

	err = cmd.Wait()
	BuildProcessHaveExit = true

	if err != nil {
		log.Println("Wait Failed")
		log.Println(err)
	}

	os.Chdir(this.currentPath)

	return err
}

func checkTimer(cmd *exec.Cmd, comp *Compile, id int) {
	for {
		// if building process hava exit normally, exit timer
		if BuildProcessHaveExit {
			log.Println("Building Process Exit Normally.")
			return
		}

		stn := time.Now()
		now := stn.UnixNano()
		// over 10s
		if now-BuildStartTime > 10*1000000000 {
			comp.buildOverTime = true
			log.Println("Building Out of Time, Terminated!")
			cmd.Process.Kill()

			systemTag := com.SubString(runtime.GOOS, 0, 5)
			if systemTag == "linux" {
				// ps -ef|grep cc1|grep 5.c|awk '{print $2}'|xargs kill -9
				cleanScript := fmt.Sprintf("ps -ef|grep cc1|grep %d.c|awk '{print $2}'|xargs kill -9", id)
				cleanCmd := exec.Command("sh",
					"-c",
					cleanScript,
				)
				err := cleanCmd.Run()
				if err != nil {
					log.Println("clean orphan failed")
				}
			}
			return
		}
	}
}
