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
  // We no longer embed the API key in the URL to keep it out of shell history.
  // The user will run 'gorenel login' after installation.

  if (opts.os === 'windows') {
    return `iwr -useb ${base}/install.ps1 -OutFile install.ps1; Get-Content install.ps1; .\\install.ps1; gorenel login`;
  }

  return `curl -fsSL ${base}/install.sh -o install.sh && cat install.sh && bash install.sh && gorenel login`;
}

/** Shorter line when CLI is already installed and in PATH. */
export function tunnelQuickCommandMinimal(opts: {
  apiKey: string;
  os: TunnelQuickOs;
  hostname: string;
}): string {
  const srv = connectServerSuffix(opts.hostname);
  const sep = opts.os === 'windows' ? ';' : '&&';
  // Use interactive login instead of 'config set' with plain text key
  return `gorenel login ${sep} gorenel connect --port 3000${srv}`;
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
  return new URL(path, base).toString();
}
