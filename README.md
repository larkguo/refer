
## Golang

[Golang标准库文档](https://studygolang.com/pkgdoc)

[Go语言中文网——Go语言标准库](https://books.studygolang.com/The-Golang-Standard-Library-by-Example/)


[Go by Example 中文](https://books.studygolang.com/gobyexample/)

[Go by Example 图解数组](https://zhuanlan.zhihu.com/p/82209614)


[100行http(s)代理](https://studygolang.com/articles/11967)


[IO模型图](https://mojotv.cn/tutorial/golang-interface-reader-writer)

[io库分析-rot13Reader-copyBuffer](https://pmlpml.github.io/ServiceComputingOnCloud/oo-thinking-abstract.html)

[IO妙用](https://zhuanlan.zhihu.com/p/26783694)

[自己实现Reader和Writer](https://www.jianshu.com/p/6bda40d003b4)


[Go文件操作大全](https://colobu.com/2016/10/12/go-file-operations/)


[全面总结:Golang调用C/C++,例子式教程. Go三种方式调用C](https://cloud.tencent.com/developer/article/1343141)

[Golang调C的so动态库和a静态库](https://jermine.vdo.pub/go/golang%E4%B8%8Ec%E4%BA%92%E7%9B%B8%E8%B0%83%E7%94%A8%E4%BB%A5%E5%8F%8A%E8%B0%83%E7%94%A8c%E7%9A%84so%E5%8A%A8%E6%80%81%E5%BA%93%E5%92%8Ca%E9%9D%99%E6%80%81%E5%BA%93/)


[cpu等系统信息库](https://github.com/shirou/gopsutil)


[使用 Go 协程和通道实现一个工作池 ](https://books.studygolang.com/gobyexample/worker-pools/)

[Go并发调度器解析之实现一个高性能协程池, 含Goroutine Pool模型图](https://zhuanlan.zhihu.com/p/37754274)

[How to Install Go](https://go-repo.io/)


## Glusterfs

[版本](https://github.com/gluster/glusterdocs/blob/master/docs/release-notes/index.md)


[gluster glusterfs glusterd glusterfsd区别](https://www.jianshu.com/p/a33ff57f32df)
cli(gluster) -> server(glusterd->glusterfsd)
client(glusterfs) -> server(glusterd->glusterfsd)
app/nfs/smba client(libgfapi) -> server(glusterd->glusterfsd)

[几个translator中继功能](https://www.tuicool.com/articles/neIVJf)

[glusterfs编译](https://github.com/gluster/glusterdocs/blob/master/docs/Developer-guide/Building-GlusterFS.md)

[4.1.10编译日志](https://github.com/larkguo/refer/blob/master/glusterfs/build/build.log)


[server-world安装使用](https://www.server-world.info/en/note?os=CentOS_7&p=glusterfs&f=1)

[volume修复-迁移-均衡](https://github.com/meetbill/op_practice_book/blob/master/doc/store/glusterfs.md)

[TStor Samba-NFS-iSCSI](https://github.com/maqingqing/TStor/wiki)


[gluster命令行日志](https://github.com/larkguo/refer/blob/master/glusterfs/log/var-log-glusterfs/cli.log-20190828)

[glusterd管理模块日志](https://github.com/larkguo/refer/blob/master/glusterfs/log/var-log-glusterfs/glusterd.log)

[glusterfs客户端distributed日志](https://github.com/larkguo/refer/blob/master/glusterfs/log/var-log-glusterfs/glusterfs-mnt.log-20190828)

[glusterfs客户端replica日志](https://github.com/larkguo/refer/blob/master/glusterfs/log/var-log-glusterfs/glusterfs-replica.log-20190828)

[glusterfsd服务端distributed日志](https://github.com/larkguo/refer/blob/master/glusterfs/log/var-log-glusterfs/bricks/glusterfs-distributed.log.1566439506)

[glusterfsd服务端replica日志](https://github.com/larkguo/refer/blob/master/glusterfs/log/var-log-glusterfs/bricks/glusterfs-replica.log-20190903)


[架构图](https://github.com/gluster/glusterdocs/tree/master/docs/images)


[故障处理](https://github.com/gluster/glusterdocs/tree/master/docs/Troubleshooting)

[故障处理-zh](https://gluster-cn.readthedocs.io/zh_CN/latest/Administrator%20Guide/Troubleshooting/)

[Peer Rejected (Connected)故障处理](https://gluster-documentations.readthedocs.io/en/master/Administrator%20Guide/Resolving%20Peer%20Rejected/)


[gogfapi源码](https://github.com/gluster/gogfapi)

[gogfapi编译及实例](https://github.com/larkguo/refer/blob/master/glusterfs/build/build.log)

[Gluster libgfapi接口和应用实例](https://blog.csdn.net/liuaigui/article/details/38443357)

[gogfapi的一个封装](https://github.com/prashanthpai/antbird)


[x-Archive智能云归档](http://www.taocloudx.com/index.php?a=shows&catid=4&id=112)

[glusterfs资源](https://blog.csdn.net/liuaigui/article/details/17331557)




## Zfs

[Quick Start Guide](https://www.freebsd.org/doc/handbook/zfs-quickstart.html)

[zfs-0.7-release源码](https://github.com/zfsonlinux/zfs/tree/zfs-0.7-release)

[Oracle Solaris管理:ZFS文件系统-管理ZFS存储池属性](https://docs.oracle.com/cd/E26926_01/html/E25826/gfifk.html)

  systemctl preset zfs-import-cache zfs-import-scan zfs-import.target zfs-mount zfs-share zfs-zed zfs.target 
[升级到zfs-0.7.4发行版时,建议用户手动重置zfs systemd预设,否则,可能导致重新引导系统时池无法自动导入.](https://github.com/Greek64/zfs/wiki/RHEL-and-CentOS)


## MegaCli 管理工具

[raid0热插拔](https://github.com/meetbill/op_practice_book/wiki/megacli02)



## iSCSI

[使用iSCSI服务部署网络存储](https://www.w3cschool.cn/linuxprobe/linuxprobe-t8nm3259.html)



## Docker

[docker release notes](https://docs.docker.com/engine/release-notes/)

[docker二进制稳定版本](https://download.docker.com/linux/static/stable/x86_64/)

[官方二进制安装 Install Docker Engine - Community from binaries](https://docs.docker.com/install/linux/docker-ce/binaries/)

[docker 18.09.0二进制安装](https://www.cnblogs.com/xiaochina/p/10469715.html)


[官方yum安装,清空容器使用命令 rm -rf /var/lib/docker](https://docs.docker.com/install/linux/docker-ce/centos/)

[linux下如何使用docker二进制文件安装](https://www.linuxprobe.com/linux-docker-biner.html)

[server-world rpm安装默认版本](https://www.server-world.info/en/note?os=CentOS_7&p=docker&f=1)

[Centos7 安装指定版本的 Docker](https://blog.51cto.com/michaelkang/2391894)

[docker安装日志](https://github.com/larkguo/refer/blob/master/docker/docker-install.log)

[daocloud的docker一键安装](http://get.daocloud.io/)

[docker源码](https://github.com/docker/engine/releases)

[docker daemon故障处理 Configure and troubleshoot the Docker daemon](https://docs.docker.com/config/daemon/)

[Docker中文资源](http://www.docker.org.cn/page/resources.html)


## 对象存储

[MinIO文档库](https://docs.min.io/cn/minio-docker-quickstart-guide.html)

[S3存储类型](https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/dev/storage-class-intro.html)

[S3客户端源码及报文](https://github.com/larkguo/refer/blob/master/s3/client/)

[对象生命周期管理](https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/dev/object-lifecycle-mgmt.html)

[Amazon S3 Glacier 入门](https://docs.aws.amazon.com/zh_cn/amazonglacier/latest/dev/amazon-glacier-getting-started.html)

[适用于 Go 的 AWS 开发工具包](https://aws.amazon.com/sdk-for-go/)
