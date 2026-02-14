import { useState } from 'react';
import { ChevronRight, ChevronLeft, Check, Shield, Zap, Bot, Moon, Sun, Bell, Server } from 'lucide-react';
import type { UserPreferences, ApprovalMode, Theme } from '../../types';

interface OnboardingWizardProps {
  isOpen: boolean;
  onComplete: (preferences: UserPreferences) => void;
}

type Step = 'welcome' | 'approval' | 'theme' | 'notifications' | 'mcp' | 'complete';

const STEPS: Step[] = ['welcome', 'approval', 'theme', 'notifications', 'complete'];

export function OnboardingWizard({ isOpen, onComplete }: OnboardingWizardProps) {
  const [currentStep, setCurrentStep] = useState<Step>('welcome');
  const [preferences, setPreferences] = useState<UserPreferences>({
    apiKey: '',
    authMethod: 'anthropic-api',
    approvalMode: 'suggest',
    defaultModel: 'sonnet',
    theme: 'dark',
    notificationsEnabled: true,
    mcpServers: [],
    onboardingCompleted: true,
  });

  if (!isOpen) return null;

  const currentIndex = STEPS.indexOf(currentStep);
  const progress = ((currentIndex + 1) / STEPS.length) * 100;

  const handleNext = () => {
    const nextIndex = currentIndex + 1;
    if (nextIndex < STEPS.length) {
      setCurrentStep(STEPS[nextIndex]);
    }
  };

  const handleBack = () => {
    const prevIndex = currentIndex - 1;
    if (prevIndex >= 0) {
      setCurrentStep(STEPS[prevIndex]);
    }
  };

  const handleComplete = () => {
    onComplete(preferences);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950">
      <div className="w-full max-w-xl p-8">
        {/* Progress bar */}
        <div className="mb-8">
          <div className="h-1 bg-slate-800 rounded-full overflow-hidden">
            <div
              className="h-full bg-blue-500 transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
          <div className="flex justify-between mt-2 text-xs text-slate-500">
            <span>Step {currentIndex + 1} of {STEPS.length}</span>
            <span>{Math.round(progress)}% complete</span>
          </div>
        </div>

        {/* Content */}
        <div className="min-h-[400px]">
          {currentStep === 'welcome' && (
            <WelcomeStep onNext={handleNext} />
          )}
          {currentStep === 'approval' && (
            <ApprovalStep
              value={preferences.approvalMode}
              onChange={(mode) => setPreferences({ ...preferences, approvalMode: mode })}
              onNext={handleNext}
              onBack={handleBack}
            />
          )}
          {currentStep === 'theme' && (
            <ThemeStep
              value={preferences.theme}
              onChange={(theme) => setPreferences({ ...preferences, theme })}
              onNext={handleNext}
              onBack={handleBack}
            />
          )}
          {currentStep === 'notifications' && (
            <NotificationsStep
              value={preferences.notificationsEnabled}
              onChange={(enabled) => setPreferences({ ...preferences, notificationsEnabled: enabled })}
              onNext={handleNext}
              onBack={handleBack}
            />
          )}
          {currentStep === 'complete' && (
            <CompleteStep
              onComplete={handleComplete}
              onBack={handleBack}
            />
          )}
        </div>
      </div>
    </div>
  );
}

// Welcome Step
function WelcomeStep({ onNext }: { onNext: () => void }) {
  return (
    <div className="text-center">
      <div className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-blue-500/20 flex items-center justify-center">
        <Bot className="w-10 h-10 text-blue-500" />
      </div>
      <h1 className="text-2xl font-bold text-slate-100 mb-3">
        Welcome to Boatman
      </h1>
      <p className="text-slate-400 mb-8 max-w-md mx-auto">
        Your desktop companion for Claude Code. Let's set up a few things to get you started.
      </p>
      <button
        onClick={onNext}
        className="flex items-center gap-2 px-6 py-3 mx-auto bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
      >
        Get Started
        <ChevronRight className="w-4 h-4" />
      </button>
    </div>
  );
}

// Approval Mode Step
function ApprovalStep({
  value,
  onChange,
  onNext,
  onBack,
}: {
  value: ApprovalMode;
  onChange: (mode: ApprovalMode) => void;
  onNext: () => void;
  onBack: () => void;
}) {
  const modes: { value: ApprovalMode; label: string; description: string; icon: React.ReactNode }[] = [
    {
      value: 'suggest',
      label: 'Suggest Mode',
      description: 'Claude suggests changes, you approve everything. Best for learning.',
      icon: <Shield className="w-6 h-6" />,
    },
    {
      value: 'auto-edit',
      label: 'Auto-Edit Mode',
      description: 'Claude can edit files automatically, but asks for commands.',
      icon: <Zap className="w-6 h-6" />,
    },
    {
      value: 'full-auto',
      label: 'Full Auto Mode',
      description: 'Claude has full control. For experienced users only.',
      icon: <Bot className="w-6 h-6" />,
    },
  ];

  return (
    <div>
      <h2 className="text-xl font-bold text-slate-100 mb-2">
        Choose Your Approval Mode
      </h2>
      <p className="text-slate-400 mb-6">
        How much control should Claude have?
      </p>

      <div className="space-y-3 mb-8">
        {modes.map((mode) => (
          <button
            key={mode.value}
            onClick={() => onChange(mode.value)}
            className={`w-full flex items-start gap-4 p-4 rounded-lg border text-left transition-colors ${
              value === mode.value
                ? 'border-blue-500 bg-blue-500/10'
                : 'border-slate-700 hover:border-slate-600'
            }`}
          >
            <div
              className={`p-2 rounded-lg ${
                value === mode.value
                  ? 'bg-blue-500/20 text-blue-500'
                  : 'bg-slate-800 text-slate-400'
              }`}
            >
              {mode.icon}
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <span className="font-medium text-slate-100">{mode.label}</span>
                {mode.value === 'suggest' && (
                  <span className="text-xs px-2 py-0.5 bg-blue-500/20 text-blue-500 rounded-full">
                    Recommended
                  </span>
                )}
              </div>
              <p className="text-sm text-slate-400 mt-1">{mode.description}</p>
            </div>
            {value === mode.value && (
              <Check className="w-5 h-5 text-blue-500 flex-shrink-0" />
            )}
          </button>
        ))}
      </div>

      <div className="flex justify-between">
        <button
          onClick={onBack}
          className="flex items-center gap-2 px-4 py-2 text-slate-400 hover:text-slate-200 transition-colors"
        >
          <ChevronLeft className="w-4 h-4" />
          Back
        </button>
        <button
          onClick={onNext}
          className="flex items-center gap-2 px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          Continue
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}

// Theme Step
function ThemeStep({
  value,
  onChange,
  onNext,
  onBack,
}: {
  value: Theme;
  onChange: (theme: Theme) => void;
  onNext: () => void;
  onBack: () => void;
}) {
  return (
    <div>
      <h2 className="text-xl font-bold text-slate-100 mb-2">
        Choose Your Theme
      </h2>
      <p className="text-slate-400 mb-6">
        Select your preferred appearance
      </p>

      <div className="flex gap-4 mb-8">
        <button
          onClick={() => onChange('dark')}
          className={`flex-1 flex flex-col items-center gap-3 p-6 rounded-lg border transition-colors ${
            value === 'dark'
              ? 'border-blue-500 bg-blue-500/10'
              : 'border-slate-700 hover:border-slate-600'
          }`}
        >
          <div className="w-16 h-16 rounded-xl bg-slate-800 flex items-center justify-center">
            <Moon className="w-8 h-8 text-slate-300" />
          </div>
          <span className="font-medium text-slate-100">Dark</span>
          {value === 'dark' && (
            <Check className="w-5 h-5 text-blue-500" />
          )}
        </button>
        <button
          onClick={() => onChange('light')}
          className={`flex-1 flex flex-col items-center gap-3 p-6 rounded-lg border transition-colors ${
            value === 'light'
              ? 'border-blue-500 bg-blue-500/10'
              : 'border-slate-700 hover:border-slate-600'
          }`}
        >
          <div className="w-16 h-16 rounded-xl bg-slate-200 flex items-center justify-center">
            <Sun className="w-8 h-8 text-slate-800" />
          </div>
          <span className="font-medium text-slate-100">Light</span>
          {value === 'light' && (
            <Check className="w-5 h-5 text-blue-500" />
          )}
        </button>
      </div>

      <div className="flex justify-between">
        <button
          onClick={onBack}
          className="flex items-center gap-2 px-4 py-2 text-slate-400 hover:text-slate-200 transition-colors"
        >
          <ChevronLeft className="w-4 h-4" />
          Back
        </button>
        <button
          onClick={onNext}
          className="flex items-center gap-2 px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          Continue
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}

// Notifications Step
function NotificationsStep({
  value,
  onChange,
  onNext,
  onBack,
}: {
  value: boolean;
  onChange: (enabled: boolean) => void;
  onNext: () => void;
  onBack: () => void;
}) {
  return (
    <div>
      <h2 className="text-xl font-bold text-slate-100 mb-2">
        Enable Notifications?
      </h2>
      <p className="text-slate-400 mb-6">
        Get notified when tasks complete or require your attention
      </p>

      <div className="flex gap-4 mb-8">
        <button
          onClick={() => onChange(true)}
          className={`flex-1 flex flex-col items-center gap-3 p-6 rounded-lg border transition-colors ${
            value
              ? 'border-blue-500 bg-blue-500/10'
              : 'border-slate-700 hover:border-slate-600'
          }`}
        >
          <div className="w-16 h-16 rounded-xl bg-blue-500/20 flex items-center justify-center">
            <Bell className="w-8 h-8 text-blue-500" />
          </div>
          <span className="font-medium text-slate-100">Yes, notify me</span>
          {value && <Check className="w-5 h-5 text-blue-500" />}
        </button>
        <button
          onClick={() => onChange(false)}
          className={`flex-1 flex flex-col items-center gap-3 p-6 rounded-lg border transition-colors ${
            !value
              ? 'border-blue-500 bg-blue-500/10'
              : 'border-slate-700 hover:border-slate-600'
          }`}
        >
          <div className="w-16 h-16 rounded-xl bg-slate-800 flex items-center justify-center">
            <Bell className="w-8 h-8 text-slate-500" />
          </div>
          <span className="font-medium text-slate-100">No thanks</span>
          {!value && <Check className="w-5 h-5 text-blue-500" />}
        </button>
      </div>

      <div className="flex justify-between">
        <button
          onClick={onBack}
          className="flex items-center gap-2 px-4 py-2 text-slate-400 hover:text-slate-200 transition-colors"
        >
          <ChevronLeft className="w-4 h-4" />
          Back
        </button>
        <button
          onClick={onNext}
          className="flex items-center gap-2 px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          Continue
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}

// Complete Step
function CompleteStep({
  onComplete,
  onBack,
}: {
  onComplete: () => void;
  onBack: () => void;
}) {
  return (
    <div className="text-center">
      <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-green-500/20 flex items-center justify-center">
        <Check className="w-10 h-10 text-green-500" />
      </div>
      <h2 className="text-2xl font-bold text-slate-100 mb-3">
        You're all set!
      </h2>
      <p className="text-slate-400 mb-8 max-w-md mx-auto">
        Your preferences have been saved. You can change them anytime in Settings.
      </p>
      <div className="flex justify-center gap-4">
        <button
          onClick={onBack}
          className="flex items-center gap-2 px-4 py-2 text-slate-400 hover:text-slate-200 transition-colors"
        >
          <ChevronLeft className="w-4 h-4" />
          Back
        </button>
        <button
          onClick={onComplete}
          className="flex items-center gap-2 px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          Start Using Boatman
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}
