@echo off
echo building...

set pan=%~d0
set current_path=%~dp0
cd %current_path%
cd ../
set root_path=%cd%

echo root path=%root_path%
echo current path=%current_path%

cd %current_path%
set bin_path=%root_path%\bin
set pkg_path=%root_path%\pkg
set vendor_path=%root_path%\vendor

::添加环境变量,即在原来的环境变量后加上英文状态下的分号和路径
set GOPATH=%vendor_path%;%root_path%

::如果pkg目录存在，直接删除
if exist %pkg_path% (
 rd %pkg_path% /S /Q
 )

::如果vendor目录不存在，创建该目录
if not exist %vendor_path% (
 md %vendor_path%
 md %vendor_path%\src
 cd %root_path%
 call "%root_path%\bin\vendor_install.bat"
 cd %current_path%
)


cd %root_path%\src
::cp -rf ${root_path}"/web/login.html" ${root_path}"/web/dist/"
xcopy  %root_path%\web\login.html %root_path%\web\dist\ /y /q
statik -f -src=%root_path%\web\dist"

cd %current_path%

::cd %pan%
::进入当前目录
::cd %current_path%
::build构建项目
call go build -p 4 -race wing-crontab
::install安装
call go install wing-crontab
::删除根目录下的可执行文件
del %root_path%\wing-crontab.exe

echo build success
echo %root_path%\bin\wing-crontab.exe