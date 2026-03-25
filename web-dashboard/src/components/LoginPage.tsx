import React, { useState } from 'react';
import { Lock, Mail, Loader2, ArrowRight, Github, Languages } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { api } from '../api/client';
import { AuthLayout } from './AuthLayout';
import { Input } from './ui/Input';
import { Button } from './ui/Button';
import { Alert } from './ui/Alert';

interface LoginPageProps {
    onLoginSuccess: (user: any) => void;
    onSwitchToRegister: () => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ onLoginSuccess, onSwitchToRegister }) => {
    const { t, i18n } = useTranslation();
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
                window.location.href = data.redirect_url;
            } else if (data.user) {
                onLoginSuccess(data.user);
            }
        } catch (err: any) {
            const data = err.response?.data;
            const errorMessage = data?.error?.message || data?.error?.Message || data?.message || data?.Message || data?.error || 'Access Denied. Please verify your credentials.';
            setError(typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage));
        } finally {
            setLoading(false);
        }
    };

    const handleSocialLogin = async (provider: string) => {
        setLoading(true);
        try {
            const data = await api.socialLogin(provider);
            if (data.redirect_url) {
                window.location.href = data.redirect_url;
            }
        } catch (err: any) {
            setError('Social login failed. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    const toggleLanguage = () => {
        const newLang = i18n.language === 'en' ? 'tr' : 'en';
        i18n.changeLanguage(newLang);
    };

    return (
        <AuthLayout 
            title={t('landing.title')} 
            subtitle={t('landing.subtitle')}
            topRight={
                <button
                    onClick={toggleLanguage}
                    className="flex items-center gap-1.5 px-2.5 py-1.5 bg-white/[0.04] border border-white/[0.08] rounded-lg text-[11px] font-medium hover:bg-white/[0.07] transition-all"
                    type="button"
                >
                    <Languages size={12} className="text-emerald-400/70" />
                    {i18n.language.toUpperCase()}
                </button>
            }
        >
            <div className="space-y-6">
                <div className="space-y-1.5">
                    <h2 className="text-xl font-semibold tracking-tight text-white">Sign in</h2>
                    <p className="text-sm text-white/40">Use your credentials or continue with SSO.</p>
                </div>

                <form onSubmit={handleSubmit} className="space-y-5">
                    {error && (
                        <Alert variant="error" title="Access denied">{error}</Alert>
                    )}

                    <div className="space-y-3.5">
                        <div className="space-y-1.5">
                            <label className="text-xs font-medium text-white/40 ml-0.5">
                                {t('auth.email')}
                            </label>
                            <Input
                                type="email"
                                required
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                placeholder="you@example.com"
                                leftIcon={<Mail className="w-4 h-4" />}
                            />
                        </div>

                        <div className="space-y-1.5">
                            <label className="text-xs font-medium text-white/40 ml-0.5">
                                {t('auth.password')}
                            </label>
                            <Input
                                type="password"
                                required
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="••••••••"
                                leftIcon={<Lock className="w-4 h-4" />}
                            />
                        </div>
                    </div>

                    <Button
                        type="submit"
                        disabled={loading}
                        variant="primary"
                        size="lg"
                        className="w-full"
                    >
                        {loading ? (
                            <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                            <>
                                {t('landing.cta')}
                                <ArrowRight className="w-4 h-4" />
                            </>
                        )}
                    </Button>
                </form>

                <div className="relative">
                    <div className="absolute inset-0 flex items-center">
                        <div className="w-full border-t border-white/[0.06]"></div>
                    </div>
                    <div className="relative flex justify-center text-xs">
                        <span className="bg-[#0d0f14] px-3 text-white/25 font-medium">or continue with</span>
                    </div>
                </div>

                <div className="grid grid-cols-2 gap-3">
                    <Button
                        onClick={() => handleSocialLogin('google')}
                        disabled={loading}
                        variant="secondary"
                        size="md"
                        type="button"
                        className="w-full justify-center"
                    >
                        <svg className="w-4 h-4 text-white/40" viewBox="0 0 24 24">
                            <path fill="currentColor" d="M12.48 10.92v3.28h7.84c-.24 1.84-.9 3.24-2.04 4.38-1.26 1.26-3.26 2.4-6.48 2.4-5.06 0-9.14-4.12-9.14-9.18s4.08-9.18 9.14-9.18c2.82 0 4.92 1.1 6.36 2.4l2.4-2.4C18.54 1.08 15.72 0 12.48 0 5.61 0 0 5.61 0 12.48s5.61 12.48 12.48 12.48c3.75 0 6.6-1.23 8.79-3.54 2.19-2.31 2.88-5.52 2.88-8.19 0-.63-.06-1.26-.15-1.89H12.48z" />
                        </svg>
                        Google
                    </Button>
                    <Button
                        onClick={() => handleSocialLogin('github')}
                        disabled={loading}
                        variant="secondary"
                        size="md"
                        type="button"
                        className="w-full justify-center"
                    >
                        <Github className="w-4 h-4 text-white/40" />
                        GitHub
                    </Button>
                </div>

                <p className="text-center text-[13px] text-white/40">
                    New here?{' '}
                    <button
                        onClick={onSwitchToRegister}
                        type="button"
                        className="text-emerald-400 hover:text-emerald-300 font-medium transition-colors"
                    >
                        Create an account
                    </button>
                </p>
            </div>
        </AuthLayout>
    );
};
