import { useEffect, useCallback, useState } from 'react';
import { useStore } from '../store';
import type { UserPreferences } from '../types';

// Import Wails bindings
import {
  GetPreferences,
  SetPreferences,
  IsOnboardingCompleted,
  CompleteOnboarding,
} from '../../wailsjs/go/main/App';

export function usePreferences() {
  const {
    preferences,
    setPreferences,
    updatePreferences,
    onboardingOpen,
    setOnboardingOpen,
  } = useStore();

  const [isLoading, setIsLoading] = useState(true);

  // Load preferences on mount
  useEffect(() => {
    const loadPreferences = async () => {
      try {
        // Check if Wails bindings are available
        if (typeof GetPreferences !== 'function') {
          console.warn('Wails bindings not available yet');
          setIsLoading(false);
          return;
        }

        const prefs = await GetPreferences();
        setPreferences(prefs as unknown as UserPreferences);

        // Check if onboarding is needed
        const completed = await IsOnboardingCompleted();
        if (!completed) {
          setOnboardingOpen(true);
        }
      } catch (err) {
        console.error('Failed to load preferences:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadPreferences();
  }, [setPreferences, setOnboardingOpen]);

  // Save preferences
  const savePreferences = useCallback(async (prefs: UserPreferences) => {
    try {
      await SetPreferences(prefs as any);
      setPreferences(prefs);
    } catch (err) {
      console.error('Failed to save preferences:', err);
    }
  }, [setPreferences]);

  // Complete onboarding
  const completeOnboardingFlow = useCallback(async (prefs: UserPreferences) => {
    try {
      await SetPreferences(prefs as any);
      await CompleteOnboarding();
      setPreferences(prefs);
      setOnboardingOpen(false);
    } catch (err) {
      console.error('Failed to complete onboarding:', err);
    }
  }, [setPreferences, setOnboardingOpen]);

  return {
    preferences,
    onboardingOpen,
    isLoading,
    savePreferences,
    updatePreferences,
    completeOnboardingFlow,
    setOnboardingOpen,
  };
}
