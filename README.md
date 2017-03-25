# github-windows-installer

GitHub Windows 安装器，简称 GWI。

## 问题简介

GitHub Windows 是在线安装的，需要连接亚马逊云。因为你懂的原因，使得安装 GitHub Windows 成了一个问题 :sob: 

## 解决方案

本库是一个 GitHub Windows 安装器的 golang 实现，在 _网络条件好的地方_ 运行就可以制作安装包啦！

网络条件好的地方：国外服务器。比如阿里云**按量付费**的 ECS，选硅谷节点最低配置。

## 使用步骤

1. 在国外服务器上下载或构建 gwi
2. 运行 gwi，将在当前工作目录生成 github-windows.zip 安装包 
3. 下载安装包到本地后运行 GitHub.application
4. 安装完成！

## 原理

1. 下载应用元数据文件
2. 下载包描述文件
3. 解析所需包/资源文件下载路径
4. 并发下载

具体请看代码 :smirk:
