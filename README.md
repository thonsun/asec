# develop install env

0.linux go env install
```shell script
wget https://golang.google.cn/dl/go1.14.9.linux-arm64.tar.gz

# .profile
export GOROOT=/usr/local/go
export GOPATH=～/Documents/GoprojectLinux  #路径自定义，多GOPATH以 ： 分割
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOROOT/bin:$GOBIN
source .profile
```

1.postgresql >=9.5

1.1 install
> 在Ubuntu下安装Postgresql后，会自动注册为服务，并随操作系统自动启动会自动添加一个名为postgres的操作系统用户，密码是随机的。并且会自动生成一个名字为postgres的数据库，用户名也为postgres，密码也是随机的。
```shell script
sudo apt-get install postgresql-9.5
sudo -u postgres psql # sudo -u postgres 是使用postgres 用户登录的意思
sudo passwd -d postgres
sudo -u postgres passwd # thonsun 
```

1.2 setup database
```sql
create user asec with password 'asec'; -- revoke all on database postgres from test;drop role username
create database asec owner asec; -- drop database dbname
grant all privileges on database asec to asec;
```

1.3 edit postgresql config
```shell script
vim /etc/postgresql/9.5/main/pg_hba.conf
```