import React, { useState } from 'react';
import { X, Download, Copy, Check, Terminal, Command } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from './ui/Button';

interface ConnectModalProps {
    isOpen: boolean;
    onClose: () => void;
    apiKey?: string;
}

export const ConnectModal: React.FC<ConnectModalProps> = ({ isOpen, onClose, apiKey }) => {
    const { t } = useTranslation();
    const [copied, setCopied] = useState(false);

    if (!isOpen) return null;

    const host = window.location.hostname;
    const isWindows = typeof window !== 'undefined' && /win/i.test(navigator.platform);
    const apiToken = apiKey || 'demo-key-12345';
    const binary = isWindows ? 'gorenel' : 'gorenel';
    
    const serverUrl = host === 'localhost' ? 'ws://localhost:9091' : `wss://${host}/tunnel/connect`;
    const command = `${binary} config set api_key ${apiToken} && ${binary} connect --server ${serverUrl} --port 3000`;

    const handleCopy = () => {
        navigator.clipboard.writeText(command);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
            <div
                className="bg-[#0d0f14] border border-white/[0.08] w-full max-w-lg rounded-2xl shadow-modal overflow-hidden animate-scale-in relative"
                role="dialog"
                aria-modal="true"
                aria-label={t('connect_modal.title')}
            >
                <button
                    onClick={onClose}
                    className="absolute top-4 right-4 p-1.5 hover:bg-white/[0.06] rounded-lg transition-all text-white/40 hover:text-white"
                    aria-label={t('common.close', 'Close')}
                >
                    <X className="w-4 h-4" />
                </button>

                <div className="p-7 text-center space-y-7">
                    <div className="space-y-2">
                        <div className="w-12 h-12 bg-emerald-500/10 rounded-xl flex items-center justify-center mx-auto border border-emerald-500/15">
                            <Terminal className="w-6 h-6 text-emerald-400" />
                        </div>
                        <h2 className="text-lg font-semibold text-white tracking-tight">{t('connect_modal.title')}</h2>
                        <p className="text-sm text-white/40">{t('connect_modal.subtitle')}</p>
                    </div>

                    <div className="space-y-2.5">
                        <a
                            href={t('connect_modal.download_url')}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="group flex items-center justify-center gap-2.5 w-full py-3.5 bg-white text-[#080a10] font-semibold rounded-xl transition-all duration-200 active:scale-[0.98] shadow-[0_1px_3px_rgba(0,0,0,0.2),0_8px_24px_rgba(0,0,0,0.25)] hover:bg-white/95 text-sm"
                        >
                            <Download className="w-4 h-4" />
                            {t('connect_modal.download_btn')}
                        </a>
                        <p className="text-[11px] text-white/25">v1.0.1 &middot; 64-bit &middot; Standalone Binary</p>
                    </div>

                    <div className="space-y-2.5 pt-3 border-t border-white/[0.04]">
                        <div className="flex items-center justify-center gap-2 text-sm text-white/60">
                            <Command className="w-3.5 h-3.5 text-emerald-400/70" />
                            <span className="font-medium">{t('connect_modal.command_label')}</span>
                        </div>
                        <div
                            onClick={handleCopy}
                            className="bg-black/30 border border-white/[0.06] rounded-xl p-4 flex items-center gap-3 cursor-pointer group hover:border-white/[0.12] transition-all text-left"
                        >
                            <code className="text-white/70 font-mono text-xs flex-1 break-all leading-relaxed">
                                {command}
                            </code>
                            <div className="p-1.5 bg-white/[0.04] rounded-lg group-hover:bg-emerald-500/15 transition-colors shrink-0">
                                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5 text-white/50 group-hover:text-emerald-400" />}
                            </div>
                        </div>
                        <p className="text-[11px] text-white/25">{t('connect_modal.command_footer')}</p>
                    </div>

                    <Button type="button" variant="outline" size="md" className="w-full" onClick={onClose}>
                        Close
                    </Button>
                </div>
            </div>
        </div>
    );
};
