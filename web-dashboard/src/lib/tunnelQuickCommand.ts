/** One-liners for install + API key + connect (dashboard copy-paste). */

export type TunnelQuickOs = 'windows' | 'unix';

function installBaseUrl(hostname: string, protocol: string): string {
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'https://gorenel.site';
  }
  return `${protocol}//${hostname}`;
}

function connectServerSuffix(hostname: string): string {
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return ' --server ws://localhost:9091';
  }
  return '';
}

/** Full line: install script + save key + connect (default for new users). */
export function tunnelQuickCommandFull(opts: {
  apiKey: string;
  os: TunnelQuickOs;
  hostname: string;
  protocol: string;
}): string {
  const base = installBaseUrl(opts.hostname, opts.protocol);
  const srv = connectServerSuffix(opts.hostname);

  if (opts.os === 'windows') {
    return `iwr -useb ${base}/install.ps1 | iex; $g = "$env:LOCALAPPDATA\\gorenel\\gorenel.exe"; & $g config set api_key ${opts.apiKey}; & $g connect --port 3000${srv}`;
  }

  const bin = '"$HOME/.gorenel/bin/gorenel"';
  return `curl -sSL ${base}/install.sh | bash && ${bin} config set api_key ${opts.apiKey} && ${bin} connect --port 3000${srv}`;
}

/** Shorter line when CLI is already installed (PATH not required on Windows). */
export function tunnelQuickCommandMinimal(opts: {
  apiKey: string;
  os: TunnelQuickOs;
  hostname: string;
}): string {
  const srv = connectServerSuffix(opts.hostname);

  if (opts.os === 'windows') {
    // Standardizing on a cleaner format that is easier to read/edit if needed
    return `$g = "$env:LOCALAPPDATA\\gorenel\\gorenel.exe"; & $g config set api_key ${opts.apiKey}; & $g connect --port 3000${srv}`;
  }

  const bin = '"$HOME/.gorenel/bin/gorenel"';
  return `${bin} config set api_key ${opts.apiKey} && ${bin} connect --port 3000${srv}`;
}

/** URL for "Magic Install" download with API key. */
export function tunnelMagicDownloadUrl(opts: {
  apiKey: string;
  os: TunnelQuickOs;
  hostname: string;
  protocol: string;
}): string {
  const base = installBaseUrl(opts.hostname, opts.protocol);
  const path = opts.os === 'windows' ? '/install.ps1' : '/install.sh';
  const url = new URL(path, base);
  url.searchParams.set('api_key', opts.apiKey);
  return url.toString();
}
