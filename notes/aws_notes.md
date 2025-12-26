# set up
aws runs on linux, we must set GOOS=linux

on windows:
powershell:
$Env:GOOS="linux"
go build main.go