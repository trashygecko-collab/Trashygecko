@Echo off
del /f /s /q "%HOMEPATH%\Lokale Einstellungen\Temp\*.*"
del /f /s /q "%HOMEPATH%\Lokale Einstellungen\Temporary Internet Files\*.*"
del /f /s /q "%HOMEPATH%\Lokale Einstellungen\Verlauf\*.*"
del /f /s /q "%windir%\Temp\*.*"
rmdir /s /q "%HOMEPATH%\Lokale Einstellungen\Temp\"

ipconfig /flushdns
pause