#define AppName "SyncTray"
#define AppExe "synctray.exe"
#define TaskName "SyncTray"
#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif

[Setup]
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher=sujaybokil
AppId={{A1B2C3D4-E5F6-7890-ABCD-EF1234567890}
VersionInfoVersion={#AppVersion}
LicenseFile=LICENSE

; Install to %LOCALAPPDATA%\SyncTray — no admin rights needed
DefaultDirName={localappdata}\SyncTray
UsePreviousAppDir=yes
DisableDirPage=yes
DisableProgramGroupPage=yes
PrivilegesRequired=lowest

; Close a running previous version before [Files] replaces it. Inno Setup uses
; Windows Restart Manager and will stop on a lock it cannot safely close.
CloseApplications=yes
CloseApplicationsFilter={#AppExe}
RestartApplications=no

OutputDir=Output
OutputBaseFilename=synctray-setup
SetupIconFile=icon.ico
UninstallDisplayIcon={app}\icon.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern

[Files]
Source: "{#AppExe}"; DestDir: "{app}"; Flags: ignoreversion
Source: "icon.ico";  DestDir: "{app}"; Flags: ignoreversion skipifsourcedoesntexist

[Icons]
Name: "{autoprograms}\{#AppName}"; Filename: "{app}\{#AppExe}"; WorkingDir: "{app}"; IconFilename: "{app}\icon.ico"

[Run]
; Offer to launch the updated application now.
Filename: "{app}\{#AppExe}"; \
  Description: "Start {#AppName} now"; \
  Flags: postinstall nowait skipifsilent

[UninstallRun]
; Remove the logon task
Filename: "{sys}\schtasks.exe"; Parameters: "/delete /tn ""{#TaskName}"" /f"; \
  Flags: runhidden; RunOnceId: "DeleteSyncTrayTask"

; Kill the running process so files can be deleted
Filename: "taskkill.exe"; Parameters: "/f /im {#AppExe}"; \
  Flags: runhidden; RunOnceId: "StopSyncTray"

[UninstallDelete]
Type: files; Name: "{userstartup}\{#AppName}.lnk"

[Code]
procedure CreateSyncTrayConfig();
var
  ConfigPath: String;
  Content: String;
begin
  ConfigPath := ExpandConstant('{app}\synctray.conf');
  if FileExists(ConfigPath) then
    exit;

  Content := '# SyncTray settings' + #13#10 +
    '# Leave empty to find syncthing.exe on PATH at every start.' + #13#10 +
    '# Set a full path only when Syncthing is not on PATH.' + #13#10 +
    'syncthing=' + #13#10 +
    '# webui=http://127.0.0.1:8384' + #13#10 +
    '# folder=C:\Path\To\Your\SyncFolder' + #13#10;
  SaveStringToFile(ConfigPath, Content, False);
end;

function CreateLogonTask(): Boolean;
var
  ResultCode: Integer;
  Parameters: String;
  ElevatedParameters: String;
  UserName: String;
begin
  Log('Creating the SyncTray user logon task.');
  Parameters := '/create /tn "{#TaskName}" /tr """' +
    ExpandConstant('{app}\{#AppExe}') + '""" /sc ONLOGON /rl LIMITED /delay 0000:30 /it /f';
  Result := Exec(ExpandConstant('{sys}\schtasks.exe'), Parameters,
    '', SW_HIDE, ewWaitUntilTerminated, ResultCode) and (ResultCode = 0);
  if Result then
  begin
    Log('Created the SyncTray user logon task.');
    exit;
  end;

  Log(Format('User-level SyncTray logon-task creation failed (exit code %d).', [ResultCode]));
  if MsgBox('Windows denied creation of the SyncTray logon task. Allow an administrator prompt to retry?',
    mbConfirmation, MB_YESNO) <> IDYES then
    exit;

  UserName := GetEnv('USERNAME');
  if GetEnv('USERDOMAIN') <> '' then
    UserName := GetEnv('USERDOMAIN') + '\' + UserName;
  ElevatedParameters := '/create /tn "{#TaskName}" /tr """' +
    ExpandConstant('{app}\{#AppExe}') + '""" /sc ONLOGON /ru "' +
    UserName + '" /rl LIMITED /delay 0000:30 /it /f';
  Result := ShellExec('runas', ExpandConstant('{sys}\schtasks.exe'), ElevatedParameters,
    '', SW_HIDE, ewWaitUntilTerminated, ResultCode) and (ResultCode = 0);
  if Result then
    Log('Created the SyncTray user logon task after elevation.')
  else
    Log(Format('Elevated SyncTray logon-task creation failed (exit code %d).', [ResultCode]));
end;

procedure CreateStartupFallback();
begin
  Log('Creating the SyncTray Startup-folder fallback shortcut.');
  CreateShellLink(ExpandConstant('{userstartup}\{#AppName}.lnk'),
    'Start SyncTray at logon', ExpandConstant('{app}\{#AppExe}'), '',
    ExpandConstant('{app}'), ExpandConstant('{app}\icon.ico'), 0, SW_SHOWNORMAL);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    Log('Running SyncTray post-install actions.');
    CreateSyncTrayConfig();
    DeleteFile(ExpandConstant('{userstartup}\{#AppName}.lnk'));
    if not CreateLogonTask() then
      CreateStartupFallback();
  end;
end;
