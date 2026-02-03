; Beot Windows Installer Script
; Inno Setup 6.x
; https://jrsoftware.org/isinfo.php

#define MyAppName "Beot"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "Beot Project"
#define MyAppURL "https://github.com/yourusername/beot"
#define MyAppExeName "beot.exe"

[Setup]
; App identity
AppId={{8F3E4B2A-1D5C-4A7F-9E8B-6C2D4F5A3B1E}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases

; Installation settings
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
PrivilegesRequiredOverridesAllowed=dialog
OutputDir=..\dist
OutputBaseFilename=BeotSetup-{#MyAppVersion}
SetupIconFile=..\assets\beot.ico
UninstallDisplayIcon={app}\{#MyAppExeName}

; Compression
Compression=lzma2
SolidCompression=yes

; Windows version requirements
MinVersion=10.0

; Modern wizard style
WizardStyle=modern
WizardSizePercent=100

; License and info (optional - uncomment if you have these files)
; LicenseFile=..\LICENSE
; InfoBeforeFile=..\README.md

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "addtopath"; Description: "Add Beot to PATH (recommended)"; GroupDescription: "Additional options:"; Flags: checkedonce
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "..\dist\beot.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; Start Menu shortcut - opens terminal and runs beot
Name: "{group}\{#MyAppName}"; Filename: "{cmd}"; Parameters: "/k ""{app}\{#MyAppExeName}"""; WorkingDir: "{userdocs}"; Comment: "Start a Beot Pomodoro session"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
; Desktop shortcut (optional)
Name: "{autodesktop}\{#MyAppName}"; Filename: "{cmd}"; Parameters: "/k ""{app}\{#MyAppExeName}"""; WorkingDir: "{userdocs}"; Tasks: desktopicon; Comment: "Start a Beot Pomodoro session"

[Code]
const
  EnvironmentKey = 'Environment';

function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
  ParamExpanded: string;
begin
  ParamExpanded := ExpandConstant(Param);
  if not RegQueryStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Uppercase(ParamExpanded) + ';', ';' + Uppercase(OrigPath) + ';') = 0;
end;

procedure AddToPath();
var
  OrigPath: string;
  NewPath: string;
  AppDir: string;
begin
  AppDir := ExpandConstant('{app}');

  if not RegQueryStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', OrigPath) then
    OrigPath := '';

  // Check if already in path
  if Pos(';' + Uppercase(AppDir) + ';', ';' + Uppercase(OrigPath) + ';') > 0 then
    exit;

  // Add to path
  if OrigPath <> '' then
    NewPath := OrigPath + ';' + AppDir
  else
    NewPath := AppDir;

  RegWriteStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', NewPath);

  // Notify Windows of environment change
  // This broadcasts WM_SETTINGCHANGE so new terminals pick up the change
end;

procedure RemoveFromPath();
var
  OrigPath: string;
  NewPath: string;
  AppDir: string;
  P: Integer;
begin
  AppDir := ExpandConstant('{app}');

  if not RegQueryStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', OrigPath) then
    exit;

  NewPath := OrigPath;

  // Remove the app directory from path (handle various positions)
  // At the beginning with semicolon after
  P := Pos(Uppercase(AppDir) + ';', Uppercase(NewPath));
  if P > 0 then
  begin
    Delete(NewPath, P, Length(AppDir) + 1);
  end
  else
  begin
    // In the middle or at the end with semicolon before
    P := Pos(';' + Uppercase(AppDir), Uppercase(NewPath));
    if P > 0 then
    begin
      Delete(NewPath, P, Length(AppDir) + 1);
    end
    else
    begin
      // Only entry (no semicolons)
      if Uppercase(NewPath) = Uppercase(AppDir) then
        NewPath := '';
    end;
  end;

  // Clean up any double semicolons
  while Pos(';;', NewPath) > 0 do
    StringChangeEx(NewPath, ';;', ';', True);

  // Remove leading/trailing semicolons
  if (Length(NewPath) > 0) and (NewPath[1] = ';') then
    Delete(NewPath, 1, 1);
  if (Length(NewPath) > 0) and (NewPath[Length(NewPath)] = ';') then
    Delete(NewPath, Length(NewPath), 1);

  RegWriteStringValue(HKEY_CURRENT_USER, EnvironmentKey, 'Path', NewPath);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    if IsTaskSelected('addtopath') then
      AddToPath();
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    RemoveFromPath();
  end;
end;

[Run]
; Option to launch after install
Filename: "{cmd}"; Parameters: "/k ""{app}\{#MyAppExeName}"""; WorkingDir: "{userdocs}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent
