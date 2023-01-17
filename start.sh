#!/bin/sh
#REM *************************************************************************
#REM 启动脚本
#REM ***************************************************************************

echo "starting..."


#先执行stop命令
SHELL_DIR=`dirname $0`
sh $SHELL_DIR/stop.sh


nohup ./main -test=false -pre=6   &

echo "start success"
