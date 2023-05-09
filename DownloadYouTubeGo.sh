# /etc/cron.d/ytdl
# 
# go run TEST-Go.go
go run /opt/DownloadYouTubeGo/DownloadYouTubeGo-1.14/DownloadYouTubeGo.go  >> /proc/1/fd/1;
echo "DONE"  >> /proc/1/fd/1;