# zpmd

github 事件通知处理


## 使用方法

    ./zpmd -install
    
    systemctl start zpmd

创建脚本文件 /var/zpmd/task.sh， 脚本内容根据实际情况编写， 比如监控 nginx 服务

    NGINX=`systemctl status nginx  | grep Active | awk '{print $3}' | cut -d "(" -f2 | cut -d ")" -f1`
    
    if [ "$NGINX" = "running" ]
    then
        echo "nginx is running"
    else
        systemctl start nginx
    fi

创建github项目触发脚本， /var/zpmd/<项目名称>_push.sh, 程序在收到github push通知时会自动执行该脚本， 

除了push 其他事件也可以支持， 比如 issues 使用脚本/var/zpmd/<项目名称>_issues.sh


## github webhook 配置

    http://host:2029/notify
    