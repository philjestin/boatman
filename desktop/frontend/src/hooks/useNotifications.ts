import { useEffect, useCallback } from 'react';
import { useStore } from '../store';

// Import Wails bindings
import { SendNotification } from '../../wailsjs/go/main/App';

export function useNotifications() {
  const { preferences } = useStore();

  const sendNotification = useCallback(async (title: string, message: string) => {
    if (!preferences?.notificationsEnabled) return;

    try {
      // Use native notification via Wails
      await SendNotification(title, message);
    } catch (err) {
      console.error('Failed to send notification:', err);
    }
  }, [preferences?.notificationsEnabled]);

  const notifyTaskComplete = useCallback((taskName: string) => {
    sendNotification('Task Complete', `"${taskName}" has been completed.`);
  }, [sendNotification]);

  const notifyApprovalRequired = useCallback((actionDescription: string) => {
    sendNotification('Approval Required', actionDescription);
  }, [sendNotification]);

  const notifyError = useCallback((errorMessage: string) => {
    sendNotification('Error', errorMessage);
  }, [sendNotification]);

  return {
    sendNotification,
    notifyTaskComplete,
    notifyApprovalRequired,
    notifyError,
  };
}
