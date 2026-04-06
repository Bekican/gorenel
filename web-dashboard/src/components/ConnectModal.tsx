import React, { useMemo, useState } from 'react';
import { X, Download, Copy, Check, Terminal, Sparkles, Zap, ShieldCheck } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from './ui/Button';
import {
  tunnelQuickCommandFull,
  tunnelQuickCommandMinimal,
  tunnelMagicDownloadUrl,
  type TunnelQuickOs,
} from '../lib/tunnelQuickCommand';

interface ConnectModalProps {
  isOpen: boolean;
  onClose: () => void;
  apiKey?: string;
}

export const ConnectModal: React.FC<ConnectModalProps> = ({ isOpen, onClose, apiKey }) => {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const [osTab, setOsTab] = useState<TunnelQuickOs>(() =>
    typeof navigator !== 'undefined' && /win/i.test(navigator.platform || '') ? 'windows' : 'unix',
  );
  const [minimal, setMinimal] = useState(false);
  const [installMode, setInstallMode] = useState<'magic' | 'manual'>('magic');

  const apiToken = apiKey || 'YOUR_API_KEY';

  const command = useMemo(() => {
    if (typeof window === 'undefined') return '';
    const { hostname, protocol } = window.location;
    if (minimal) {
      return tunnelQuickCommandMinimal({ apiKey: apiToken, os: osTab, hostname });
    }
    return tunnelQuickCommandFull({ apiKey: apiToken, os: osTab, hostname, protocol });
  }, [apiToken, osTab, minimal]);

  const downloadHref = useMemo(() => {
    if (typeof window === 'undefined') return '/downloads/gorenel-windows-amd64.exe';
    const { hostname, protocol } = window.location;
    if (installMode === 'magic') {
      return tunnelMagicDownloadUrl({ apiKey: apiToken, os: osTab, hostname, protocol });
    }
    const o = window.location.origin;
    return osTab === 'windows' ? `${o}/downloads/gorenel-windows-amd64.exe` : `${o}/install.sh`;
  }, [osTab, apiToken, installMode]);

  if (!isOpen) return null;

  const handleCopy = () => {
    navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const missingKey = !apiKey;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div
        className="bg-[#0d0f14] border border-white/[0.08] w-full max-w-lg rounded-2xl shadow-modal overflow-hidden animate-scale-in relative max-h-[90dvh] flex flex-col"
        role="dialog"
        aria-modal="true"
        aria-label={t('connect_modal.title')}
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute top-4 right-4 p-1.5 hover:bg-white/[0.06] rounded-lg transition-all text-white/40 hover:text-white"
          aria-label={t('common.close', 'Close')}
        >
          <X className="w-4 h-4" />
        </button>

        <div className="p-6 md:p-7 space-y-5 overflow-y-auto">
          <div className="text-center space-y-2 pr-8">
            <div className="w-12 h-12 bg-emerald-500/10 rounded-xl flex items-center justify-center mx-auto border border-emerald-500/15">
              <Sparkles className="w-6 h-6 text-emerald-400" />
            </div>
            <h2 className="text-lg font-semibold text-white tracking-tight">{t('connect_modal.title')}</h2>
            <p className="text-sm text-white/45 leading-relaxed">{t('connect_modal.subtitle')}</p>
          </div>

          {missingKey && (
            <p className="rounded-xl border border-amber-500/20 bg-amber-500/[0.06] px-3 py-2 text-center text-xs text-amber-200/90">
              {t('connect_modal.need_key')}
            </p>
          )}

          <div className="flex rounded-lg border border-white/[0.08] bg-black/20 p-0.5">
            <button
              type="button"
              onClick={() => setInstallMode('magic')}
              className={`flex-1 rounded-md py-2.5 text-xs font-semibold flex items-center justify-center gap-2 transition ${
                installMode === 'magic' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'text-white/35 hover:text-white/60 border border-transparent'
              }`}
            >
              <Zap size={13} />
              {t('connect_modal.mode_magic')}
            </button>
            <button
              type="button"
              onClick={() => setInstallMode('manual')}
              className={`flex-1 rounded-md py-2.5 text-xs font-semibold flex items-center justify-center gap-2 transition ${
                installMode === 'manual' ? 'bg-white/[0.08] text-white border border-white/[0.1]' : 'text-white/35 hover:text-white/60 border border-transparent'
              }`}
            >
              <Terminal size={13} />
              {t('connect_modal.mode_manual')}
            </button>
          </div>

          {installMode === 'magic' ? (
            <div className="space-y-6 py-2">
              <div className="space-y-4">
                <div className="flex rounded-lg border border-white/[0.08] bg-black/20 p-0.5">
                  <button
                    type="button"
                    onClick={() => setOsTab('windows')}
                    className={`flex-1 rounded-md py-2 text-xs font-medium transition ${
                      osTab === 'windows' ? 'bg-white/[0.1] text-white' : 'text-white/40 hover:text-white/65'
                    }`}
                  >
                    Windows
                  </button>
                  <button
                    type="button"
                    onClick={() => setOsTab('unix')}
                    className={`flex-1 rounded-md py-2 text-xs font-medium transition ${
                      osTab === 'unix' ? 'bg-white/[0.1] text-white' : 'text-white/40 hover:text-white/65'
                    }`}
                  >
                    Linux / Mac
                  </button>
                </div>

                <div className="space-y-3">
                  <div className="flex items-center gap-2 text-xs font-medium text-white/55">
                    <Terminal className="w-3.5 h-3.5 text-emerald-400/80" />
                    {t('connect_modal.magic_label')}
                  </div>
                  <button
                    type="button"
                    onClick={handleCopy}
                    className="w-full bg-black/35 border border-emerald-500/20 rounded-xl p-4 flex items-start gap-3 text-left hover:bg-emerald-500/[0.03] transition-all group"
                  >
                    <code className="text-emerald-400 font-mono text-[11px] md:text-xs flex-1 break-all leading-relaxed">
                      {command}
                    </code>
                    <div className="p-2 bg-emerald-500/10 rounded-lg group-hover:bg-emerald-500/20 transition-colors shrink-0 mt-0.5">
                      {copied ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4 text-emerald-400/60 group-hover:text-emerald-400" />}
                    </div>
                  </button>
                  <p className="text-[10px] text-white/30 leading-relaxed italic">
                    {osTab === 'windows' 
                      ? t('connect_modal.magic_footer_win')
                      : t('connect_modal.magic_footer_unix')}
                  </p>
                </div>
              </div>

              <div className="p-3 bg-blue-500/[0.03] border border-blue-500/10 rounded-xl flex items-center gap-3">
                <ShieldCheck className="w-4 h-4 text-blue-400/60" />
                <p className="text-[10px] text-blue-300/40 font-medium leading-normal">
                   {t('connect_modal.security_note')}
                </p>
              </div>
            </div>
          ) : (
            <div className="space-y-5">
              <div className="flex rounded-lg border border-white/[0.08] bg-black/20 p-0.5">
                <button
                  type="button"
                  onClick={() => setOsTab('windows')}
                  className={`flex-1 rounded-md py-2 text-xs font-medium transition ${
                    osTab === 'windows' ? 'bg-white/[0.1] text-white' : 'text-white/40 hover:text-white/65'
                  }`}
                >
                  {t('connect_modal.tab_windows')}
                </button>
                <button
                  type="button"
                  onClick={() => setOsTab('unix')}
                  className={`flex-1 rounded-md py-2 text-xs font-medium transition ${
                    osTab === 'unix' ? 'bg-white/[0.1] text-white' : 'text-white/40 hover:text-white/65'
                  }`}
                >
                  {t('connect_modal.tab_unix')}
                </button>
              </div>

              <label className="flex cursor-pointer items-center gap-2.5 text-xs text-white/50">
                <input
                  type="checkbox"
                  checked={minimal}
                  onChange={(e) => setMinimal(e.target.checked)}
                  className="rounded border-white/20 bg-black/40 text-emerald-500 focus:ring-emerald-500/30"
                />
                {t('connect_modal.minimal_toggle')}
              </label>

              <div className="space-y-2">
                <div className="flex items-center gap-2 text-xs font-medium text-white/55">
                  <Terminal className="w-3.5 h-3.5 text-emerald-400/80" />
                  {minimal ? t('connect_modal.command_label_short') : t('connect_modal.command_label')}
                </div>
                <button
                  type="button"
                  onClick={handleCopy}
                  className="w-full bg-black/35 border border-white/[0.08] rounded-xl p-3.5 flex items-start gap-3 text-left hover:border-emerald-500/20 transition-colors group"
                >
                  <code className="text-white/75 font-mono text-[11px] md:text-xs flex-1 break-all leading-relaxed">
                    {command}
                  </code>
                  <div className="p-1.5 bg-white/[0.05] rounded-lg group-hover:bg-emerald-500/15 transition-colors shrink-0 mt-0.5">
                    {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5 text-white/45 group-hover:text-emerald-400" />}
                  </div>
                </button>
                <p className="text-[11px] text-white/30 leading-relaxed">{t('connect_modal.command_footer')}</p>
              </div>

              <a
                href={downloadHref}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center justify-center gap-2 w-full py-2.5 text-xs font-medium text-white/45 hover:text-emerald-300/90 border border-white/[0.06] rounded-xl hover:border-white/[0.1] transition-colors"
              >
                <Download className="w-3.5 h-3.5" />
                {osTab === 'windows' ? t('connect_modal.download_only_win') : t('connect_modal.download_only_unix')}
              </a>
            </div>
          )}

          <Button type="button" variant="outline" size="md" className="w-full" onClick={onClose}>
            {t('common.close', 'Close')}
          </Button>
        </div>
      </div>
    </div>
  );
};
