三台consul服务群集搭建

###初始化脚本
````
#!/usr/bin/env bash      
echo "正在初始化目录..."      
mkdir /usr/local/consul      
mkdir /usr/local/consul/data      
mkdir /usr/local/consul/conf      
mkdir /usr/local/consul/conf/server      
mkdir /usr/local/consul/run      
      
cd /usr/local/consul/run      
wget https://releases.hashicorp.com/consul/1.0.1/consul_1.0.1_linux_amd64.zip?_ga=2.180601889.768715731.1512704037-1064535734.1512475239
mv consul_1.0.1_linux_amd64.zip?_ga=2.180601889.768715731.1512704037-1064535734.1512475239 consul.zip
unzip consul.zip      
rm -rf consul.zip      
ll 
````     
以上初始化脚本在每台服务器上执行      
      
###环境准备，三台虚拟机
192.168.152.131    
192.168.152.132    
192.168.152.133    
    
第一台配置文件如下（192.168.152.131）    
````      
{      
    "datacenter": "dc1",      
    "data_dir": "/usr/local/consul/data",      
    "log_level": "INFO",      
    "node_name": "consul-node-1",      
    "server": true,      
    "ui": true,      
    "bootstrap": true,      
    "bind_addr": "192.168.152.131",      
    "client_addr": "192.168.152.131",      
    "retry_interval": "3s",      
    "raft_protocol": 3,      
    "enable_debug": false,      
    "rejoin_after_leave": true,      
    "disable_update_check": true,      
    "enable_debug": true      
}      
````
写入文件/usr/local/consul/conf/server/config.json      
启动服务      
./run/consul agent -config-dir ./conf/server -pid-file=./run/consul-server.pid      
      
第二台配置文件如下（192.168.152.132） 
````     
{      
    "datacenter": "dc1",      
    "data_dir": "/usr/local/consul/data",      
    "log_level": "INFO",      
    "node_name": "consul-node-2",      
    "server": true,      
    "ui": true,      
    "bind_addr": "192.168.152.132",      
    "client_addr": "192.168.152.132",      
    "retry_join": [      
        "192.168.152.131",      
        "192.168.152.132",      
        "192.168.152.133"      
    ],      
    "start_join": [      
        "192.168.152.131",      
        "192.168.152.132",      
        "192.168.152.133"      
    ],      
    "retry_interval": "3s",      
    "raft_protocol": 3,      
    "rejoin_after_leave": true,      
    "disable_update_check": true,      
    "enable_debug": true      
}      
````      
写入文件/usr/local/consul/conf/server/config.json      
启动服务   
````         
./run/consul agent -config-dir ./conf/server -pid-file=./run/consul-server.pid      
````            
第三台配置文件如下（192.168.152.133）    
````  
{      
    "datacenter": "dc1",      
    "data_dir": "/usr/local/consul/data",      
    "log_level": "INFO",      
    "node_name": "consul-node-3",      
    "server": true,      
    "ui": true,      
    "bind_addr": "192.168.152.133",      
    "client_addr": "192.168.152.133",      
    "retry_join": [      
        "192.168.152.131",      
        "192.168.152.132",      
        "192.168.152.133"      
    ],      
    "start_join": [      
        "192.168.152.131",      
        "192.168.152.132",      
        "192.168.152.133"      
    ],      
    "retry_interval": "3s",      
    "raft_protocol": 3,      
    "rejoin_after_leave": true,      
    "disable_update_check": true,      
    "enable_debug": true      
}    
````  
写入文件/usr/local/consul/conf/server/config.json      
启动服务      
````
./run/consul agent -config-dir ./conf/server -pid-file=./run/consul-server.pid      
````  
可选操作，关闭防火墙 
````     
systemctl status firewalld      
systemctl stop firewalld  
````    
查看集群成员
````
./run/consul members --http-addr 192.168.152.131:8500
Node           Address               Status  Type    Build  Protocol  DC   Segment
consul-node-1  192.168.152.131:8301  alive   server  1.0.1  2         dc1  <all>
consul-node-2  192.168.152.132:8301  alive   server  1.0.1  2         dc1  <all>
consul-node-3  192.168.152.133:8301  alive   server  1.0.1  2         dc1  <all>
````
查询集群leader，可以强制退出一台机器，查看leader的分配情况
````
curl 192.168.152.131:8500/v1/status/leader
````
退出192.168.152.131，多试几次得到新分类的leader
````
curl 192.168.152.132:8500/v1/status/leader
"192.168.152.133:8300"
````
dns查询
````
dig @192.168.152.131 -p 8600 consul-node-1.node.consul
````
其中的consul-node-1为node名称


重新加载配置文件
````
consul reload --http-addr=192.168.152.131:8500
````

访问http://192.168.152.131:8500可以看到consul的问管理界面


接下来我们在192.168.152.131上安装nginx，用来做负载均衡    
http://blog.51cto.com/cantgis/1540004    
添加 Nginx源[资源库]
````    
sudo rpm -Uvh http://nginx.org/packages/centos/7/noarch/RPMS/nginx-release-centos-7-0.el7.ngx.noarch.rpm    
````
安装 Nginx    
````
sudo yum install nginx    
````
开启 Nginx服务    
````
sudo systemctl start nginx.service    
````
或者 
````   
/usr/sbin/nginx -c /etc/nginx/nginx.conf    
ps aux | grep nginx    
[root@localhost consul]# ps aux| grep nginx
root       4414  0.0  0.0  46296   952 ?        Ss   05:32   0:00 nginx: master process /usr/sbin/nginx -c /etc/nginx/nginx.conf
nginx      4415  0.0  0.1  48768  1968 ?        S    05:32   0:00 nginx: worker process
root       4436  0.0  0.0 112648   956 pts/1    R+   05:33   0:00 grep --color=auto nginx
[root@localhost consul]#
````
访问http://192.168.152.131/出现nginx的欢迎界面

配置负载均衡
````
upstream consul.com {
      server 192.168.152.131:8500;
      server 192.168.152.132:8500;
      server 192.168.152.133:8500;
}

server {
    listen       80;
    server_name  consul.com;

    location / {
        proxy_pass          http://consul.com;
        proxy_set_header    Host            $host;
        proxy_set_header    X-Real-IP       $remote_addr;
        proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
````
指定accesslog $upstream_addr参数可以看到负载均衡请求的服务器
````
 log_format  main  '"$upstream_addr" $remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
 ````                     