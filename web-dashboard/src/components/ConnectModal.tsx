import React from 'react';
import { X, Download, Copy, Check } from 'lucide-react';

interface ConnectModalProps {
    isOpen: boolean;
    onClose: () => void;
}

export const ConnectModal: React.FC<ConnectModalProps> = ({ isOpen, onClose }) => {
    const [copied, setCopied] = React.useState(false);

    if (!isOpen) return null;

    const command = `gorenel start --server localhost:7000 --port 3000 --api-key demo-key-12345`;

    const handleCopy = () => {
        navigator.clipboard.writeText(command);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-white w-full max-w-2xl rounded-3xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="p-6 border-b border-neutral-100 flex items-center justify-between">
                    <h2 className="text-xl font-bold text-neutral-900">Connect Your Application</h2>
                    <button onClick={onClose} className="p-2 hover:bg-neutral-100 rounded-xl transition-all">
                        <X className="w-5 h-5 text-neutral-400" />
                    </button>
                </div>

                <div className="p-8 space-y-8">
                    {/* Step 1 */}
                    <div className="flex gap-4">
                        <div className="w-8 h-8 rounded-full bg-primary-50 text-primary-600 flex items-center justify-center font-bold flex-shrink-0">1</div>
                        <div className="space-y-2">
                            <h3 className="font-bold text-neutral-900">Download CLI</h3>
                            <p className="text-sm text-neutral-500 italic">Gorenel binary'sini sistemine uygun indir ve terminalden erişebileceğin bir yere koy.</p>
                            <div className="flex gap-3 pt-2">
                                <button className="flex items-center gap-2 px-4 py-2 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-black transition-all">
                                    <Download className="w-4 h-4" /> Windows (.exe)
                                </button>
                                <button className="flex items-center gap-2 px-4 py-2 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 transition-all">
                                    Mac / Linux
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Step 2 */}
                    <div className="flex gap-4">
                        <div className="w-8 h-8 rounded-full bg-primary-50 text-primary-600 flex items-center justify-center font-bold flex-shrink-0">2</div>
                        <div className="space-y-4 flex-1">
                            <h3 className="font-bold text-neutral-900">Run Connection Command</h3>
                            <div className="bg-neutral-900 rounded-2xl p-4 relative group">
                                <code className="text-primary-400 font-mono text-sm block pr-12">
                                    {command}
                                </code>
                                <button
                                    onClick={handleCopy}
                                    className="absolute right-4 top-1/2 -translate-y-1/2 p-2 bg-white/10 hover:bg-white/20 rounded-lg transition-all"
                                >
                                    {copied ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4 text-white" />}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};