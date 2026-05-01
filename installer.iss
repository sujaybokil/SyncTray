#define AppName "SyncTray"
#define AppExe "synctray.exe"
#define TaskName "SyncTray"
#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif

[Setup]
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher=synctray
VersionInfoVersion={#AppVersion}

; Install to %LOCALAPPDATA%\SyncTray — no admin rights needed
DefaultDirName={localappdata}\SyncTray
DisableDirPage=yes
DisableProgramGroupPage=yes
PrivilegesRequired=lowest

OutputDir=Output
OutputBaseFilename=synctray-setup
SetupIconFile=icon.ico
UninstallDisplayIcon={app}\{#AppExe}
Compression=lzma
SolidCompression=yes
WizardStyle=modern

[Files]
Source: "{#AppExe}";  DestDir: "{app}"; Flags: ignoreversion
Source: "icon.ico";   DestDir: "{app}"; Flags: ignoreversion skipifsourcedoesntexist

[Run]
; Kill any running instance before finishing (in case of upgrade)
Filename: "taskkill.exe"; Parameters: "/f /im {#AppExe}"; \
  Flags: runhidden; Check: IsAppRunning

; Register the logon task
Filename: "schtasks.exe"; \
  Parameters: "/create /tn ""{#TaskName}"" /tr """"""{app}\{#AppExe}"""" /sc ONLOGON /delay 0000:30 /it /f"; \
  Flags: runhidden

; Offer to launch now
Filename: "{app}\{#AppExe}"; \
  Description: "Start {#AppName} now"; \
  Flags: postinstall nowait skipifsilent

[UninstallRun]
; Remove the logon task
Filename: "schtasks.exe"; Parameters: "/delete /tn ""{#TaskName}"" /f"; \
  Flags: runhidden

; Kill the running process so files can be deleted
Filename: "taskkill.exe"; Parameters: "/f /im {#AppExe}"; \
  Flags: runhidden

[Code]
function IsAppRunning(): Boolean;
var
  ResultCode: Integer;
begin
  // Returns true if synctray.exe is in the process list
  Exec('tasklist.exe', '/fi "IMAGENAME eq {#AppExe}" /fo csv /nh',
    '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Result := (ResultCode = 0);
end;
