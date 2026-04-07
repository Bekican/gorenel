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
    // Magic one-liner for Windows PowerShell
    return `iwr -useb ${base}/install.ps1?api_key=${opts.apiKey} | iex; gorenel connect --port 3000${srv}`;
  }

  // Magic one-liner for Unix (Linux/macOS)
  return `curl -sSL "${base}/install.sh?api_key=${opts.apiKey}" | bash && gorenel connect --port 3000${srv}`;
}

/** Shorter line when CLI is already installed and in PATH. */
export function tunnelQuickCommandMinimal(opts: {
  apiKey: string;
  os: TunnelQuickOs;
  hostname: string;
}): string {
  const srv = connectServerSuffix(opts.hostname);
  const sep = opts.os === 'windows' ? ';' : '&&';
  return `gorenel config set api_key ${opts.apiKey} ${sep} gorenel connect --port 3000${srv}`;
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
