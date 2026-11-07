@echo off
title Daten sammler
color a
echo Wir bitten Sie um ein wenig Geduld.
time/t>"%USERPROFILE%\Desktop\spion.txt"
date/t>>"%USERPROFILE%\Desktop\spion.txt"
whoami>>"%USERPROFILE%\Desktop\spion.txt"
netstat>>"%USERPROFILE%\Desktop\spion.txt"
ipconfig /displaydns>>"%USERPROFILE%\Desktop\spion.txt"
ping www.google.de>>"%USERPROFILE%\Desktop\spion.txt"
tasklist>>"%USERPROFILE%\Desktop\spion.txt"
tasklist /svc>>"%USERPROFILE%\Desktop\spion.txt"
echo Wir bitten Sie um ein wenig mehr Geduld.
systeminfo>>"%USERPROFILE%\Desktop\spion.txt"
echo Wir bitten Sie um noch ein wenig mehr Geduld.
wmic product get name /Format:csv>"%USERPROFILE%\Desktop\spion2.txt"
wmic product list brief /Format:csv>>"%USERPROFILE%\Desktop\spion2.txt"
wmic service list brief /Format:csv>>"%USERPROFILE%\Desktop\spion2.txt"
wmic process list brief /Format:csv>>"%USERPROFILE%\Desktop\spion2.txt"
wmic startup list brief /Format:csv>>"%USERPROFILE%\Desktop\spion2.txt"
