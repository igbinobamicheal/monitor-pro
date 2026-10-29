; ==============================================================================
; MonitoPro V2 - Windows Endpoint Agent Inno Setup Script
; ==============================================================================
#define MyAppName "MonitoPro Endpoint Agent"
#define MyAppVersion "2.0.0"
#define MyAppPublisher "MonitoPro Inc."
#define MyAppURL "https://monitopro.com"
#define MyAppExeName "proactive-it-agent.exe"
#define MyAppId "{7B98EFA1-B2C4-4A88-9D1E-C4782A0F81D3}"

[Setup]
AppId={#MyAppId}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={commonpf}\MonitoPro\Agent
DefaultGroupName=MonitoPro
DisableProgramGroupPage=yes
OutputDir=..\..\dist\installer
OutputBaseFilename=MonitoProAgentSetup-v2.0.0
Compression=lzma2/max
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64
PrivilegesRequired=admin
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
; The V2 agent is a single, zero-dependency, statically linked Rust binary
Source: "..\..\services\agent\target\release\{#MyAppExeName}"; DestDir: "{app}"; DestName: "{#MyAppExeName}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"

[UninstallRun]
; Stop and remove the Windows Service during uninstallation
Filename: "{app}\{#MyAppExeName}"; Parameters: "--uninstall"; Flags: runhidden
Filename: "sc.exe"; Parameters: "stop ProactiveITAgent"; Flags: runhidden
Filename: "sc.exe"; Parameters: "delete ProactiveITAgent"; Flags: runhidden

[Code]
var
  lblCopying: TNewStaticText;
  lblRegService: TNewStaticText;
  lblRegDevice: TNewStaticText;
  lblStartService: TNewStaticText;
  ConfigPage: TInputQueryWizardPage;

function GetCommandLineParam(ParamName, DefaultValue: String): String;
var
  Val: String;
begin
  Val := ExpandConstant('{param:' + ParamName + '|}');
  if Val = '' then
    Result := DefaultValue
  else
    Result := Val;
end;

procedure InitializeWizard();
begin
  // Create Custom Configuration Page
  ConfigPage := CreateInputQueryPage(wpWelcome,
    'Workspace Configuration', 'MonitoPro Endpoint Enrollment',
    'Please enter your MonitoPro tenant server URL (e.g. https://companya.monitopro.com) and the one-time enrollment token.');
  
  ConfigPage.Add('Server URL (or Subdomain API):', False);
  ConfigPage.Add('Enrollment Token:', False);
  
  // Set defaults or values from command line if supplied
  ConfigPage.Values[0] := GetCommandLineParam('SERVERURL', 'https://api.monitopro.com');
  ConfigPage.Values[1] := GetCommandLineParam('REGTOKEN', '');

  if not WizardSilent then begin
    // Customize Installing Page UI
    WizardForm.ProgressGauge.Visible := False;
    WizardForm.FilenameLabel.Visible := False;
    WizardForm.StatusLabel.Visible := False;

    // Step checklist
    lblCopying := TNewStaticText.Create(WizardForm);
    lblCopying.Parent := WizardForm.InstallingPage;
    lblCopying.Top := ScaleY(40);
    lblCopying.Left := ScaleX(20);
    lblCopying.Width := WizardForm.InstallingPage.Width - ScaleX(40);
    lblCopying.Font.Name := 'Segoe UI';
    lblCopying.Font.Size := 10;
    lblCopying.Caption := '⚙ Copying binaries and dependencies...';

    lblRegService := TNewStaticText.Create(WizardForm);
    lblRegService.Parent := WizardForm.InstallingPage;
    lblRegService.Top := ScaleY(70);
    lblRegService.Left := ScaleX(20);
    lblRegService.Width := WizardForm.InstallingPage.Width - ScaleX(40);
    lblRegService.Font.Name := 'Segoe UI';
    lblRegService.Font.Size := 10;
    lblRegService.Caption := '  Registering Windows Service (ProactiveITAgent)...';
    lblRegService.Enabled := False;

    lblRegDevice := TNewStaticText.Create(WizardForm);
    lblRegDevice.Parent := WizardForm.InstallingPage;
    lblRegDevice.Top := ScaleY(100);
    lblRegDevice.Left := ScaleX(20);
    lblRegDevice.Width := WizardForm.InstallingPage.Width - ScaleX(40);
    lblRegDevice.Font.Name := 'Segoe UI';
    lblRegDevice.Font.Size := 10;
    lblRegDevice.Caption := '  Enrolling Endpoint Identity (Ed25519)...';
    lblRegDevice.Enabled := False;

    lblStartService := TNewStaticText.Create(WizardForm);
    lblStartService.Parent := WizardForm.InstallingPage;
    lblStartService.Top := ScaleY(130);
    lblStartService.Left := ScaleX(20);
    lblStartService.Width := WizardForm.InstallingPage.Width - ScaleX(40);
    lblStartService.Font.Name := 'Segoe UI';
    lblStartService.Font.Size := 10;
    lblStartService.Caption := '  Starting Windows Service...';
    lblStartService.Enabled := False;
  end;
end;

function ShouldSkipPage(PageID: Integer): Boolean;
begin
  Result := False;
  if PageID = ConfigPage.ID then begin
    if WizardSilent then
      Result := True
    else if (GetCommandLineParam('SERVERURL', '') <> '') and (GetCommandLineParam('REGTOKEN', '') <> '') then
      Result := True;
  end;
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = ConfigPage.ID then begin
    if Trim(ConfigPage.Values[0]) = '' then begin
      MsgBox('Please enter the MonitoPro Server URL.', mbError, MB_OK);
      Result := False;
    end else if Trim(ConfigPage.Values[1]) = '' then begin
      MsgBox('Please enter a valid Enrollment Token.', mbError, MB_OK);
      Result := False;
    end;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
  ServerURL: String;
  RegToken: String;
  ExePath: String;
begin
  if CurStep = ssPostInstall then begin
    ServerURL := GetCommandLineParam('SERVERURL', '');
    if ServerURL = '' then
      ServerURL := Trim(ConfigPage.Values[0]);
    if ServerURL = '' then
      ServerURL := 'https://api.monitopro.com';

    RegToken := GetCommandLineParam('REGTOKEN', '');
    if RegToken = '' then
      RegToken := Trim(ConfigPage.Values[1]);

    ExePath := ExpandConstant('{app}\{#MyAppExeName}');

    // 1. Files copied
    if not WizardSilent then begin
      lblCopying.Caption := '✔ Binaries installed';
      lblCopying.Font.Style := [fsBold];
      WizardForm.Refresh();
    end;

    // 2. Register Service
    if not WizardSilent then begin
      lblRegService.Enabled := True;
      lblRegService.Caption := '⚙ Registering Windows Service (ProactiveITAgent)...';
      WizardForm.Refresh();
    end;
    
    if Exec(ExePath, '--install', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then begin
      if not WizardSilent then begin
        if ResultCode = 0 then begin
          lblRegService.Caption := '✔ Windows Service registered';
          lblRegService.Font.Style := [fsBold];
        end else begin
          lblRegService.Caption := '❌ Registering Service (Code: ' + IntToStr(ResultCode) + ')';
        end;
      end;
    end else begin
      if not WizardSilent then begin
        lblRegService.Caption := '❌ Registering Service (Failed to execute)';
      end;
    end;
    if not WizardSilent then WizardForm.Refresh();

    // 3. Register Device (Ed25519 Enrollment)
    if not WizardSilent then begin
      lblRegDevice.Enabled := True;
      lblRegDevice.Caption := '⚙ Enrolling Endpoint Identity...';
      WizardForm.Refresh();
    end;

    if Exec(ExePath, '--register "' + ServerURL + '" "' + RegToken + '"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then begin
      if not WizardSilent then begin
        if ResultCode = 0 then begin
          lblRegDevice.Caption := '✔ Endpoint identity enrolled';
          lblRegDevice.Font.Style := [fsBold];
        end else begin
          lblRegDevice.Caption := '❌ Endpoint Enrollment failed';
        end;
      end;
    end else begin
      if not WizardSilent then begin
        lblRegDevice.Caption := '❌ Enrollment execution failed';
      end;
    end;
    if not WizardSilent then WizardForm.Refresh();

    // 4. Start Service via sc.exe
    if not WizardSilent then begin
      lblStartService.Enabled := True;
      lblStartService.Caption := '⚙ Starting Windows Service...';
      WizardForm.Refresh();
    end;

    if Exec('sc.exe', 'start ProactiveITAgent', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then begin
      if not WizardSilent then begin
        if (ResultCode = 0) or (ResultCode = 1056) then begin // 1056 = Already running
          lblStartService.Caption := '✔ Service started and running';
          lblStartService.Font.Style := [fsBold];
        end else begin
          lblStartService.Caption := '❌ Service start failed (Code: ' + IntToStr(ResultCode) + ')';
        end;
        WizardForm.Refresh();
      end;
    end else begin
      if not WizardSilent then begin
        lblStartService.Caption := '❌ Could not start service via sc.exe';
        WizardForm.Refresh();
      end;
    end;
  end;
end;
