#!/bin/sh
#REM *************************************************************************
#REM 停止脚本
#REM ***************************************************************************

PROJECT_NAME="mock"
PROJECT_JAR_NAME="main"

echo "stoping..."

PID=$(ps -ef |grep ${PROJECT_JAR_NAME} |grep -v grep |awk '{print $2}')
if [ -z "$PID" ]
then
        echo "${PROJECT_NAME} is already stopped"
else
        echo kill $PID
        kill $PID
fi

echo "stop success"

