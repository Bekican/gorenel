import React, { useState } from 'react';
import { User, Mail, Lock, Loader2, AlertCircle, ArrowRight, Github, Languages } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { AuthLayout } from './AuthLayout';
import { api } from '../api/client';

interface RegisterPageProps {
  onSwitchToLogin: () => void;
  onRegisterSuccess: (user: any) => void;
}

export const RegisterPage: React.FC<RegisterPageProps> = ({ onSwitchToLogin, onRegisterSuccess }) => {
  const { t, i18n } = useTranslation();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const data = await api.register({ name, email, password });
      if (data.user) {
        onRegisterSuccess(data.user);
      } else {
        // After manual register, usually we might need to login
        // or the server might return user object directly
        onSwitchToLogin();
      }
    } catch (err: any) {
      const data = err.response?.data;
      const errorMessage = data?.error?.message || data?.error?.Message || data?.message || data?.Message || data?.error || 'Registration failed. Please try again.';
      setError(typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage));
    } finally {
      setLoading(false);
    }
  };

  const toggleLanguage = () => {
    const newLang = i18n.language === 'en' ? 'tr' : 'en';
    i18n.changeLanguage(newLang);
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

  return (
    <AuthLayout 
      title={t('common.register_title')} 
      subtitle={t('landing.subtitle')}
    >
      <div className="absolute top-4 right-4 z-50">
        <button
          onClick={toggleLanguage}
          className="flex items-center gap-2 px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg text-[10px] font-bold hover:bg-white/10 transition-all uppercase"
        >
          <Languages size={12} className="text-emerald-400" />
          {i18n.language.toUpperCase()}
        </button>
      </div>
      <div className="space-y-8 animate-in fade-in duration-500">
        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <div className="bg-rose-500/10 border border-rose-500/20 text-rose-400 px-4 py-3 rounded-2xl flex items-center gap-3 animate-in shake-in">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <p className="text-[11px] font-bold uppercase tracking-wider">{error}</p>
            </div>
          )}

          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-white/30 uppercase tracking-[0.2em] ml-2">
                {t('auth.name')}
              </label>
              <div className="relative group">
                <User className="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-white/20 group-focus-within:text-emerald-500 transition-colors" />
                <input
                  type="text"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full bg-black/40 border border-white/5 rounded-2xl py-4 pl-12 pr-6 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/50 transition-all"
                  placeholder="John Doe"
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-bold text-white/30 uppercase tracking-[0.2em] ml-2">
                {t('auth.email')}
              </label>
              <div className="relative group">
                <Mail className="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-white/20 group-focus-within:text-emerald-500 transition-colors" />
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full bg-black/40 border border-white/5 rounded-2xl py-4 pl-12 pr-6 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/50 transition-all"
                  placeholder="operator@gorenel.site"
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-bold text-white/30 uppercase tracking-[0.2em] ml-2">
                {t('auth.password')}
              </label>
              <div className="relative group">
                <Lock className="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-white/20 group-focus-within:text-emerald-500 transition-colors" />
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full bg-black/40 border border-white/5 rounded-2xl py-4 pl-12 pr-6 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/50 transition-all"
                  placeholder="••••••••"
                />
              </div>
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="btn-primary-premium w-full py-4 rounded-xl flex items-center justify-center gap-3 active:scale-[0.98]"
          >
            {loading ? (
              <Loader2 className="w-5 h-5 animate-spin" />
            ) : (
              <>
                <span className="font-bold uppercase tracking-widest text-xs">
                  {t('landing.cta')}
                </span>
                <ArrowRight className="w-4 h-4" />
              </>
            )}
          </button>
        </form>

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-white/5"></div>
          </div>
          <div className="relative flex justify-center text-[10px] font-bold uppercase tracking-[0.3em]">
            <span className="bg-[#101217] px-4 text-white/20">Authorized Protocols</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <button
            onClick={() => handleSocialLogin('google')}
            disabled={loading}
            className="flex items-center justify-center gap-2 p-3 rounded-xl border border-white/5 bg-white/[0.02] hover:bg-white/[0.05] transition-all group disabled:opacity-50"
          >
            <svg className="w-4 h-4 text-white/40 group-hover:text-white transition-colors" viewBox="0 0 24 24">
              <path fill="currentColor" d="M12.48 10.92v3.28h7.84c-.24 1.84-.9 3.24-2.04 4.38-1.26 1.26-3.26 2.4-6.48 2.4-5.06 0-9.14-4.12-9.14-9.18s4.08-9.18 9.14-9.18c2.82 0 4.92 1.1 6.36 2.4l2.4-2.4C18.54 1.08 15.72 0 12.48 0 5.61 0 0 5.61 0 12.48s5.61 12.48 12.48 12.48c3.75 0 6.6-1.23 8.79-3.54 2.19-2.31 2.88-5.52 2.88-8.19 0-.63-.06-1.26-.15-1.89H12.48z"/>
            </svg>
            <span className="text-[10px] font-bold uppercase tracking-widest text-white/40 group-hover:text-white transition-colors">Google</span>
          </button>
          <button
            onClick={() => handleSocialLogin('github')}
            disabled={loading}
            className="flex items-center justify-center gap-2 p-3 rounded-xl border border-white/5 bg-white/[0.02] hover:bg-white/[0.05] transition-all group disabled:opacity-50"
          >
            <Github className="w-4 h-4 text-white/40 group-hover:text-white transition-colors" />
            <span className="text-[10px] font-bold uppercase tracking-widest text-white/40 group-hover:text-white transition-colors">Github</span>
          </button>
        </div>

        <p className="text-center text-[11px] text-white/30 font-medium">
          Already verified?{' '}
          <button 
            onClick={onSwitchToLogin}
            className="text-emerald-500 hover:text-emerald-400 font-bold underline-offset-4 hover:underline transition-all"
          >
            Log in to Gateway
          </button>
        </p>
      </div>
    </AuthLayout>
  );
};
