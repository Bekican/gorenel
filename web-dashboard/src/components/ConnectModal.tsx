import React, { useState } from 'react';
import { X, Download, Copy, Check, Terminal, Command } from 'lucide-react';
import { useTranslation } from 'react-i18next';

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
    // Use .\ for Windows, and ensure WebSocket URL for Fly.dev/Nginx
    const binary = isWindows ? '.\\gorenel' : 'gorenel';
    
    // In production, use the current host but with the correct WebSocket path
    // If we're on gorenel.site, the CLI default is already correct, but being explicit is safer
    const serverUrl = host === 'localhost' ? 'ws://localhost:9091' : `wss://${host}/tunnel/connect`;
    const command = `${binary} start --server ${serverUrl} --port 3000 --api-key ${apiToken}`;

    const handleCopy = () => {
        navigator.clipboard.writeText(command);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200">
            <div className="bg-[#0A0A0A] border border-white/10 w-full max-w-lg rounded-3xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200 relative">

                {/* Close Button */}
                <button
                    onClick={onClose}
                    className="absolute top-4 right-4 p-2 hover:bg-white/10 rounded-full transition-all text-white/50 hover:text-white"
                >
                    <X className="w-5 h-5" />
                </button>

                <div className="p-8 text-center space-y-8">
                    {/* Header */}
                    <div className="space-y-2">
                        <div className="w-16 h-16 bg-primary/20 rounded-2xl flex items-center justify-center mx-auto mb-4 ring-1 ring-primary/50 shadow-[0_0_30px_-10px_rgba(16,185,129,0.3)]">
                            <Terminal className="w-8 h-8 text-primary" />
                        </div>
                        <h2 className="text-2xl font-bold text-white tracking-tight">{t('connect_modal.title')}</h2>
                        <p className="text-white/50 text-sm">{t('connect_modal.subtitle')}</p>
                    </div>

                    {/* Step 1: Download */}
                    <div className="space-y-3">
                        <a
                            href={t('connect_modal.download_url')}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="group relative flex items-center justify-center gap-3 w-full py-4 bg-primary hover:bg-emerald-400 text-black font-bold rounded-xl transition-all active:scale-[0.98] shadow-lg shadow-primary/20"
                        >
                            <Download className="w-5 h-5" />
                            <span>{t('connect_modal.download_btn')}</span>
                        </a>
                        <p className="text-xs text-white/30">v1.0.1 • 64-bit • Standalone Binary</p>
                    </div>

                    {/* Step 2: Command */}
                    <div className="space-y-3 pt-4 border-t border-white/5">
                        <div className="flex items-center justify-center gap-2 text-sm font-medium text-white/70">
                            <Command className="w-4 h-4 text-primary" />
                            <span>{t('connect_modal.command_label')}</span>
                        </div>
                        <div
                            onClick={handleCopy}
                            className="bg-black/50 border border-white/10 rounded-xl p-4 flex items-center gap-3 cursor-pointer group hover:border-primary/30 transition-all text-left"
                        >
                            <code className="text-primary font-mono text-xs md:text-sm flex-1 break-all line-clamp-2">
                                {command}
                            </code>
                            <div className="p-2 bg-white/5 rounded-lg group-hover:bg-primary/20 transition-colors">
                                {copied ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4 text-white/70 group-hover:text-primary" />}
                            </div>
                        </div>
                        <p className="text-xs text-white/30">{t('connect_modal.command_footer')}</p>
                    </div>
                </div>
            </div>
        </div>
    );
};