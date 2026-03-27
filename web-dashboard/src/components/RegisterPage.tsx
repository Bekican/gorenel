import React, { useId, useState } from 'react';
import { User, Mail, Lock, Loader2, ArrowRight, Github, Languages } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { AuthLayout } from './AuthLayout';
import { api, type UserSession } from '../api/client';
import { Input } from './ui/Input';
import { Button } from './ui/Button';
import { Alert } from './ui/Alert';

interface RegisterPageProps {
  onSwitchToLogin: () => void;
  onRegisterSuccess: (user: UserSession) => void;
}

function extractErrorMessage(err: unknown): string {
  if (!err || typeof err !== 'object') return 'Registration failed. Please try again.';
  const e = err as { response?: { data?: unknown } };
  const data = e.response?.data;
  const msg =
    typeof data === 'string'
      ? data
      : (data && typeof data === 'object'
        ? (() => {
          const d = data as Record<string, unknown>;
          const errObj = (d['error'] && typeof d['error'] === 'object') ? (d['error'] as Record<string, unknown>) : undefined;
          const fromErr = errObj?.['message'] ?? errObj?.['Message'];
          const direct = d['message'] ?? d['Message'];
          const fallback = d['error'];
          const candidate = fromErr ?? direct ?? fallback;
          return typeof candidate === 'string' ? candidate : undefined;
        })()
        : undefined);
  if (typeof msg === 'string' && msg.trim() !== '') return msg;
  try {
    return typeof msg === 'string' ? msg : JSON.stringify(msg);
  } catch {
    return 'Registration failed. Please try again.';
  }
}

export const RegisterPage: React.FC<RegisterPageProps> = ({ onSwitchToLogin, onRegisterSuccess }) => {
  const { t, i18n } = useTranslation();
  const nameId = useId();
  const emailId = useId();
  const passwordId = useId();
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
        onSwitchToLogin();
      }
    } catch (err: unknown) {
      setError(extractErrorMessage(err));
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
    } catch {
      setError('Social login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout 
      title={t('common.register_title')} 
      subtitle={t('landing.subtitle')}
      topRight={
        <button
          onClick={toggleLanguage}
          className="flex items-center gap-1.5 px-2.5 py-1.5 bg-white/[0.04] border border-white/[0.08] rounded-lg text-[11px] font-medium hover:bg-white/[0.07] transition-all"
          type="button"
          aria-label={i18n.language === 'en' ? 'Switch to Turkish' : 'Switch to English'}
        >
          <Languages size={12} className="text-emerald-400/70" />
          {i18n.language.toUpperCase()}
        </button>
      }
    >
      <div className="space-y-6">
        <div className="space-y-1.5">
          <h2 className="text-xl font-semibold tracking-tight text-white">Create account</h2>
          <p className="text-sm text-white/70">Get started with API keys, tunnels, and reserved URLs.</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          {error && (
            <Alert variant="error" title="Registration failed">{error}</Alert>
          )}

          <div className="space-y-3.5">
            <div className="space-y-1.5">
              <label htmlFor={nameId} className="text-xs font-medium text-white/75 ml-0.5">
                {t('auth.name')}
              </label>
              <Input
                id={nameId}
                type="text"
                required
                autoComplete="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="John Doe"
                leftIcon={<User className="w-4 h-4" />}
              />
            </div>

            <div className="space-y-1.5">
              <label htmlFor={emailId} className="text-xs font-medium text-white/75 ml-0.5">
                {t('auth.email')}
              </label>
              <Input
                id={emailId}
                type="email"
                required
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                leftIcon={<Mail className="w-4 h-4" />}
              />
            </div>

            <div className="space-y-1.5">
              <label htmlFor={passwordId} className="text-xs font-medium text-white/75 ml-0.5">
                {t('auth.password')}
              </label>
              <Input
                id={passwordId}
                type="password"
                required
                minLength={8}
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                leftIcon={<Lock className="w-4 h-4" />}
              />
              <p className="text-[10px] text-white/55 mt-0.5 ml-0.5">Minimum 8 characters</p>
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
            <span className="bg-[#0d0f14] px-3 text-white/55 font-medium">or continue with</span>
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
            <svg className="w-4 h-4 text-white/70" aria-hidden viewBox="0 0 24 24">
              <path fill="currentColor" d="M12.48 10.92v3.28h7.84c-.24 1.84-.9 3.24-2.04 4.38-1.26 1.26-3.26 2.4-6.48 2.4-5.06 0-9.14-4.12-9.14-9.18s4.08-9.18 9.14-9.18c2.82 0 4.92 1.1 6.36 2.4l2.4-2.4C18.54 1.08 15.72 0 12.48 0 5.61 0 0 5.61 0 12.48s5.61 12.48 12.48 12.48c3.75 0 6.6-1.23 8.79-3.54 2.19-2.31 2.88-5.52 2.88-8.19 0-.63-.06-1.26-.15-1.89H12.48z"/>
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
            <Github className="w-4 h-4 text-white/70" aria-hidden />
            GitHub
          </Button>
        </div>

        <p className="text-center text-[13px] text-white/70">
          Already have an account?{' '}
          <button 
            onClick={onSwitchToLogin}
            type="button"
            className="text-emerald-400 hover:text-emerald-300 font-medium transition-colors"
          >
            Sign in
          </button>
        </p>
      </div>
    </AuthLayout>
  );
};
