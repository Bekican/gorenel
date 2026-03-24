#define MyAppName "Gorenel"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "Gorenel"
#define MyAppExeName "gorenel.exe"

[Setup]
AppId={{8BFBF74A-16D5-4A1A-A301-BB39C0A7F21C}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={localappdata}\gorenel
DefaultGroupName=Gorenel
DisableProgramGroupPage=yes
OutputDir=.
OutputBaseFilename=gorenel-setup
Compression=lzma
SolidCompression=yes
ChangesEnvironment=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "..\..\bin\gorenel-client-windows-amd64.exe"; DestDir: "{app}"; DestName: "gorenel.exe"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\Gorenel CLI"; Filename: "{cmd}"; Parameters: "/K gorenel version"

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath(ExpandConstant('{app}'))

[Code]
function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Uppercase(Param) + ';', ';' + Uppercase(OrigPath) + ';') = 0;
end;
