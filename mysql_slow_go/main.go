package main

import (
	"fmt"
	"log"
	"mysql_slow_go/utils"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {

	// 定义cli参数
	var (
		mysqlSlowLog = kingpin.Flag("log", "mysql慢日志路径").Short('h').Required().String()
		ptBin        = kingpin.Flag("pt", "pt-query-digest命令的绝对路径").Short('p').Default("/usr/bin/pt-query-digest").String()
		mT           = kingpin.Flag("mt", "默认查看区间时间").Short('m').Default("-1").Int()
		startT       = kingpin.Flag("start", "开始时间").Default().String()
		endT         = kingpin.Flag("end", "结束时间").Default().String()
	)
	kingpin.Parse()

	// 判断文件是否存在
	if f := utils.CheckFileIsExist(*mysqlSlowLog); f == false {
		log.Panicf("mysql 慢日志文件未找到")
	}

	if f := utils.CheckFileIsExist(*ptBin); f == false {
		log.Panicf("pt-query-digest 命令不存在")
	}

	// 字符串指针
	s := time.Now().AddDate(0, 0, *mT).Format("2006-01-02")
	var ptCmdExec *string
	ptCmdExec = &s
	log.Println(utils.GetAppPath())

	// 查询某段时间 24小时
	if *startT == "" && *endT == "" {

		exec := fmt.Sprintf("%s %s --output json --since \"%s 00:00:00\" --until \"%s 23:59:59\"", *ptBin, *mysqlSlowLog, s, s)
		ptCmdExec = &exec
		ptCmd, err := utils.ExecShellCmd(*ptCmdExec)

		if err != nil {
			log.Panicf("execute %s have errors->%s", *ptCmdExec, err)
		}
		if ptCmd == "" {
			fmt.Println("当前时间未匹配到慢日志信息")
			return
		}
		json := utils.SLOWLOG_JSON{}
		err = json.UnmarshalJSON([]byte(ptCmd))
		if err != nil {
			log.Panicf("Json 序列化失败 %s", err)
		}
		utils.ToHtml(json)
		return
	}

	// 根据时间区间查询
	if *startT != "" || *endT != "" {
		vv := utils.ValidateDate{StartAt: *startT, EndAt: *endT}
		err := validator.New().Struct(vv)
		if err != nil {
			fmt.Println("如果需要根据时间查询，请写 开始时间 和 结束时间 (格式: \"2020-01-02 10:26:05\")")
			return
		}
		exec := fmt.Sprintf("%s %s --output json --since \"%s\" --until \"%s\"", *ptBin, *mysqlSlowLog, *startT, *endT)
		ptCmdExec = &exec
		ptCmd, err := utils.ExecShellCmd(*ptCmdExec)
		if err != nil {
			log.Panicf("execute %s have errors->%s", *ptCmdExec, err)
		}
		if ptCmd == "" {
			fmt.Println("当前时间未匹配到慢日志信息")
			f, err := os.OpenFile("log.html", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				log.Panicf("execute to file %s", err)
			}
			_, err = f.WriteString("无慢SQL语句 - " + *startT + " --- " + *endT)
			return
		}
		json := utils.SLOWLOG_JSON{}
		err = json.UnmarshalJSON([]byte(ptCmd))
		if err != nil {
			log.Panicf("Json 序列化失败 %s", err)
		}
		utils.ToHtml(json)
		return
	}
}
