import React, { useState } from 'react';
import { Zap, Lock, Mail, Loader2, AlertCircle } from 'lucide-react';
import { api } from '../api/client';

interface LoginPageProps {
    onLoginSuccess: (user: any) => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ onLoginSuccess }) => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);

        try {
            const data = await api.login({ email, password });
            
            if (data.redirect_url) {
                // Eğer backend bir yönlendirme URL'si dönerse (SPA dostu yöntem)
                window.location.href = data.redirect_url;
            } else if (data.user) {
                onLoginSuccess(data.user);
            }
        } catch (err: any) {
            setError(err.response?.data || 'Giriş yapılamadı. Lütfen bilgilerinizi kontrol edin.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-neutral-50 flex items-center justify-center p-4">
            <div className="max-w-md w-full">
                {/* Logo & Header */}
                <div className="text-center mb-8">
                    <div className="w-16 h-16 bg-gradient-to-br from-primary-500 to-primary-600 rounded-2xl shadow-lg flex items-center justify-center mx-auto mb-4">
                        <Zap className="w-10 h-10 text-white" />
                    </div>
                    <h1 className="text-3xl font-extrabold text-neutral-900 tracking-tight">Gorenel</h1>
                    <p className="text-neutral-500 mt-2">Yönetim paneline hoş geldiniz</p>
                </div>

                {/* Form Card */}
                <div className="bg-white rounded-3xl shadow-xl shadow-neutral-200/50 border border-neutral-100 p-8">
                    <form onSubmit={handleSubmit} className="space-y-6">
                        {error && (
                            <div className="bg-red-50 border border-red-100 text-red-600 px-4 py-3 rounded-xl flex items-center gap-3 animate-shake">
                                <AlertCircle className="w-5 h-5 flex-shrink-0" />
                                <p className="text-sm font-medium">{error}</p>
                            </div>
                        )}

                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-neutral-700 ml-1">E-posta</label>
                            <div className="relative">
                                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-neutral-400" />
                                <input
                                    type="email"
                                    required
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    className="w-full bg-neutral-50 border border-neutral-200 rounded-2xl py-3.5 pl-12 pr-4 text-neutral-900 placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 transition-all"
                                    placeholder="demo@gorenel.io"
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-neutral-700 ml-1">Şifre</label>
                            <div className="relative">
                                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-neutral-400" />
                                <input
                                    type="password"
                                    required
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    className="w-full bg-neutral-50 border border-neutral-200 rounded-2xl py-3.5 pl-12 pr-4 text-neutral-900 placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 transition-all"
                                    placeholder="••••••••"
                                />
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={loading}
                            className="w-full bg-primary-600 hover:bg-primary-700 text-white font-bold py-4 rounded-2xl shadow-lg shadow-primary-500/20 active:scale-[0.98] transition-all disabled:opacity-70 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                        >
                            {loading ? (
                                <>
                                    <Loader2 className="w-5 h-5 animate-spin" />
                                    Giriş yapılıyor...
                                </>
                            ) : (
                                'Giriş Yap'
                            )}
                        </button>
                    </form>

                    <div className="mt-8 pt-6 border-t border-neutral-100 text-center">
                        <p className="text-sm text-neutral-500">
                            Demo hesabıyla giriş yapabilirsiniz: <br />
                            <span className="font-mono text-primary-600 bg-primary-50 px-2 py-0.5 rounded-md">demo@gorenel.io</span>
                        </p>
                    </div>
                </div>

                <p className="text-center text-neutral-400 text-xs mt-8">
                    Gorenel v1.0.0 • Güvenli Tünel İzleme Sistemi
                </p>
            </div>
        </div>
    );
};
