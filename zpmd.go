package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// 命令行定义
var (
	// DEBUG 可打印详细日志，包括SQL
	h         = flag.Bool("h", false, "help usage")
	debug     = flag.Bool("X", false, "run debug level")
	install   = flag.Bool("install", false, "run install")
	uninstall = flag.Bool("uninstall", false, "run uninstall")
	port      = flag.Int("p", 2029, "server listen port")
	seconds       = flag.Int("secs", 60, "task run interval seconds")
	task      = flag.String("t", "/var/zpmd/task.sh", "task script ")
)

var InstallScript = `#!/bin/bash -x
mkdir -p /var/zpmd
chmod -R 755 /var/zpmd
install -m 777 ./zpmd /usr/local/bin/zpmd
test -f /var/zpmd/cron.sh || echo "#!/bin/bash" > /var/zpmd/cron.sh 
test -d /usr/lib/systemd/system || mkdir -p /usr/lib/systemd/system
cat>/usr/lib/systemd/system/zpmd.service<<EOF
[Unit]
Description=zpmd
After=network.target

[Service]
User=root
ExecStart=/usr/local/bin/zpmd

[Install]
WantedBy=multi-user.target
EOF

chmod 600 /usr/lib/systemd/system/zpmd.service
systemctl enable zpmd && systemctl daemon-reload

`

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Install() {
	Must(ioutil.WriteFile(InstallScript, []byte("/tmp/zpmd_install.sh"), 0777))
	Must(exec.Command("/bin/bash", "/tmp/zpmd_install.sh").Run())
	Must(os.Remove("/tmp/zpmd_install.sh"))
}

func Uninstall() {
	_ = os.Remove("/usr/lib/systemd/system/zpmd.service")
	_ = os.Remove("/usr/local/bin/zpmd")
}

type Result struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

func Notify(c echo.Context) error {
	event := c.Request().Header.Get("X-GitHub-Event")
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var result Result
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	script := fmt.Sprintf("/var/zpmd/%s_%s.sh", result.Repository.Name, event)
	log.Println("bin/bash ", script)
	err = exec.Command("/bin/bash", script).Run()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"code": 0})
}


func main() {

	if *install {
		Install()
		return
	}

	if *uninstall {
		Uninstall()
		return
	}
	go func() {

		ticker := time.NewTicker(time.Second *time.Duration(*seconds))
		for t := range ticker.C {
			err := exec.Command("/bin/bash", *task).Run()
			if err != nil {
				log.Printf("%d Task run error, %s",t, err.Error())
			}
		}

	}()

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	// Init Handlers
	e.POST("/notify", Notify)
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "radius ${time_rfc3339} ${remote_ip} ${method} ${uri} ${protocol} ${status} ${id} ${user_agent} ${error}\n",
		Output: os.Stdout,
	}))
	e.HideBanner = true
	e.Debug = *debug
	log.Fatal(e.Start(fmt.Sprintf(":%d", *port)))
}
