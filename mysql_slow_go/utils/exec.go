package utils

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

// CheckFileIsExist 判断文件是否存在  存在返回 true 不存在返回false
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// ExecShellCmd 执行shell命令
func ExecShellCmd(shellcmd string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", shellcmd)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ValidateDate 验证时间字段
type ValidateDate struct {
	StartAt string `validate:"datetime=2006-01-02 15:04:05"`
	EndAt   string `validate:"datetime=2006-01-02 15:04:05"`
}

// Slowlog 输出结果
type Slowlog struct {
	Ts_min      string  // 第一次出现的时间
	Ts_max      string  // 最后一次出现的时间
	Query_count int     // 出现次数
	Query       string  // 语句示例
	Pct_95      float64 // 平均耗时(秒)
	User        string  // 用户
	Host        string  // 主机
	Db          string  // 数据库
	Query_max   float64 // 最大耗时
	Query_min   float64 // 最小耗时

}

func ToHtml(json SLOWLOG_JSON) {
	var totalRecords int
	for i, _ := range json.Classes {
		i++
		totalRecords++

	}

	slowlogs := make([]Slowlog, totalRecords)
	for i, v := range json.Classes {

		slowlogs[i].Ts_min = v.Ts_min
		slowlogs[i].Ts_max = v.Ts_max
		slowlogs[i].Host = v.Metrics.Hosts.Value
		slowlogs[i].User = v.Metrics.Users.Value
		slowlogs[i].Db = v.Metrics.Dbs.Value

		// 从JSON解析出来的Pct_95为字符串，将字符串类型转换为float64类型
		floatPct95, err := strconv.ParseFloat(v.Metrics.Query_time.Pct_95, 10)
		if err != nil {
			log.Panicf("parse pct_95 from str to float have errors->%s", err)
		}

		QueryTimeMax, err := strconv.ParseFloat(v.Metrics.Query_time.Max, 10)
		if err != nil {
			log.Panicf("parse pct_95 from str to float have errors->%s", err)
		}

		QueryTimeMin, err := strconv.ParseFloat(v.Metrics.Query_time.Min, 10)
		if err != nil {
			log.Panicf("parse pct_95 from str to float have errors->%s", err)
		}

		slowlogs[i].Pct_95 = floatPct95
		slowlogs[i].Query_count = v.Query_count
		slowlogs[i].Query = v.Examples.Query
		slowlogs[i].Query = v.Examples.Query
		slowlogs[i].Query_max = QueryTimeMax
		slowlogs[i].Query_min = QueryTimeMin
	}
	slowquery_tpl := template.New("mysql_slowquery.html")
	tmpl, err := slowquery_tpl.Parse(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <style type="text/css">
        #hor-minimalist-b
        {
            font-family: "Lucida Sans Unicode", "Lucida Grande", Sans-Serif;
            font-size: 14px;
            background: #fff;
            margin: 10px;
            width: auto;
            border-collapse: collapse;
            text-align: center;
        }
        #hor-minimalist-b th
        {
            font-size: 14px;
            font-weight: normal;
            color: #039;
            padding: 10px 8px;
            border-bottom: 2px solid #6678b1;
        }
        #hor-minimalist-b td
        {
            border-bottom: 1px solid #ccc;
            color: #669;
            padding: 6px 8px;
        }
        #hor-minimalist-b tbody tr:hover td
        {
            color: #009;
        }
    </style>
</head>
<body>
<table id="hor-minimalist-b" style="table-layout:fixed;word-break:break-all;">
    <thead>
    <tr>
        <th width="3%">序号</th>
        <th width="8.5%">数据库</th>
        <th width="10%">查询用户</th>
        <th width="49.5%">语句示例</th>
        <th width="9%">第一次出现的时间</th>
        <th width="9%">最后一次出现的时间</th>
        <th width="4.5%">出现次数</th>
        <th width="5.5%">平均耗时(秒)</th>
        <th width="5.5%">最大耗时(秒)</th>
        <th width="5.5%">最小耗时(秒)</th>
    </tr>
    </thead>
    <tbody>
    {{ range $i,$v := . }}
    <tr>
        <td title="序号" width="3%">{{ $i  }}</td>
        <td title="数据库" width="8.5%">{{ .Db }}</td>
        <td title="查询用户" width="10%">{{ .User }}@{{ .Host }}</td>
        <td title="语句示例" style="text-align: left;width: 49.5%">{{ .Query }}</td>
        <td title="第一次出现的时间" width="9.5%">{{ .Ts_min  }}</td>
        <td title="最后一次出现的时间" width="9.5%">{{ .Ts_max  }}</td>
        <td title="出现次数" width="4.5%">{{ .Query_count }}</td>
        <td title="平均耗时(秒)" width="5.5%">{{ .Pct_95 | printf "%.1f"}}</td>
        <td title="最大耗时(秒)" width="5.5%">{{ .Query_max | printf "%.1f" }}</td>
        <td title="最小耗时(秒)" width="5.5%">{{ .Query_min | printf "%.1f" }}</td>
    </tr>
    {{ end }}
    </tbody>
</table>
</body>
</html>`)

	if err != nil {
		log.Panicf("parse html file have errors->%s", err)
	}
	rendResult := new(bytes.Buffer)
	err = tmpl.Execute(rendResult, slowlogs)

	f, err := os.OpenFile("log.html", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Panicf("execute to file %s", err)
	}
	_, err = f.WriteString(rendResult.String())
	if err != nil {
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			return
		}
	}()
}

func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}
