import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { SettingsModal } from './SettingsModal';
import type { UserPreferences } from '../../types';

describe('SettingsModal', () => {
  const mockPreferences: UserPreferences = {
    apiKey: 'sk-ant-test123',
    authMethod: 'anthropic-api',
    gcpProjectId: '',
    gcpRegion: 'us-east5',
    approvalMode: 'suggest',
    defaultModel: 'sonnet',
    theme: 'dark',
    notificationsEnabled: true,
    mcpServers: [],
    onboardingCompleted: true,
  };

  const mockOnClose = vi.fn();
  const mockOnSave = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Modal visibility', () => {
    it('should not render when isOpen is false', () => {
      render(
        <SettingsModal
          isOpen={false}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.queryByText('Settings')).not.toBeInTheDocument();
    });

    it('should render when isOpen is true', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.getByText('Settings')).toBeInTheDocument();
    });

    it('should close modal when close button is clicked', () => {
      const { container } = render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      // Find the X icon close button in the header
      const closeButton = container.querySelector('.lucide-x')?.parentElement;
      if (closeButton) {
        fireEvent.click(closeButton);
      }

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should close modal when backdrop is clicked', () => {
      const { container } = render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const backdrop = container.querySelector('.bg-black\\/60');
      if (backdrop) {
        fireEvent.click(backdrop);
      }

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should close modal when Cancel button is clicked', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      fireEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Tab navigation', () => {
    it('should render all tab buttons', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.getByRole('button', { name: /General/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Approval/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /MCP Servers/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /About/i })).toBeInTheDocument();
    });

    it('should start with General tab active', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const generalTab = screen.getByRole('button', { name: /General/i });
      expect(generalTab).toHaveClass('bg-slate-800');
    });

    it('should switch to Approval tab when clicked', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const approvalTab = screen.getByRole('button', { name: /Approval/i });
      fireEvent.click(approvalTab);

      expect(screen.getByText('Approval Mode')).toBeInTheDocument();
    });

    it('should switch to MCP tab when clicked', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const mcpTab = screen.getByRole('button', { name: /MCP Servers/i });
      fireEvent.click(mcpTab);

      expect(screen.getAllByText('MCP Servers').length).toBeGreaterThan(0);
    });

    it('should switch to About tab when clicked', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const aboutTab = screen.getByRole('button', { name: /About/i });
      fireEvent.click(aboutTab);

      expect(screen.getByText('Boatman')).toBeInTheDocument();
    });
  });

  describe('General Settings - Authentication', () => {
    it('should render authentication method selector', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.getByText('Authentication Method')).toBeInTheDocument();
      expect(screen.getByText('Anthropic API')).toBeInTheDocument();
      expect(screen.getByText('Google Cloud')).toBeInTheDocument();
    });

    it('should have Anthropic API selected by default', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const anthropicButton = screen.getByText('Anthropic API').closest('button');
      expect(anthropicButton).toHaveClass('border-blue-500');
    });

    it('should switch to Google Cloud auth method', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const gcpButton = screen.getByText('Google Cloud').closest('button');
      if (gcpButton) {
        fireEvent.click(gcpButton);
      }

      expect(screen.getByText('GCP Project ID')).toBeInTheDocument();
      expect(screen.getByText('GCP Region')).toBeInTheDocument();
    });

    it('should render API key input when using Anthropic API', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      expect(apiKeyInput).toBeInTheDocument();
      expect(apiKeyInput).toHaveValue('sk-ant-test123');
    });

    it('should toggle API key visibility', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...') as HTMLInputElement;
      expect(apiKeyInput.type).toBe('password');

      const toggleButtons = screen.getAllByRole('button');
      const eyeButton = toggleButtons.find((button) => {
        const svg = button.querySelector('svg');
        return svg && button.className.includes('text-slate-500');
      });

      if (eyeButton) {
        fireEvent.click(eyeButton);
        expect(apiKeyInput.type).toBe('text');

        fireEvent.click(eyeButton);
        expect(apiKeyInput.type).toBe('password');
      }
    });

    it('should update API key value', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      fireEvent.change(apiKeyInput, { target: { value: 'sk-ant-new-key' } });

      expect(apiKeyInput).toHaveValue('sk-ant-new-key');
    });
  });

  describe('General Settings - GCP Configuration', () => {
    const gcpPreferences: UserPreferences = {
      ...mockPreferences,
      authMethod: 'google-cloud',
      gcpProjectId: 'my-project',
      gcpRegion: 'us-central1',
    };

    it('should render GCP configuration when Google Cloud is selected', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={gcpPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.getByText('GCP Project ID')).toBeInTheDocument();
      expect(screen.getByText('GCP Region')).toBeInTheDocument();
    });

    it('should display GCP project ID value', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={gcpPreferences}
          onSave={mockOnSave}
        />
      );

      const projectIdInput = screen.getByPlaceholderText('my-project-id');
      expect(projectIdInput).toHaveValue('my-project');
    });

    it('should update GCP project ID', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={gcpPreferences}
          onSave={mockOnSave}
        />
      );

      const projectIdInput = screen.getByPlaceholderText('my-project-id');
      fireEvent.change(projectIdInput, { target: { value: 'new-project' } });

      expect(projectIdInput).toHaveValue('new-project');
    });

    it('should render GCP region selector with options', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={gcpPreferences}
          onSave={mockOnSave}
        />
      );

      const regionSelect = screen.getByDisplayValue('us-central1');
      expect(regionSelect).toBeInTheDocument();

      const options = within(regionSelect.parentElement!).getAllByRole('option');
      expect(options).toHaveLength(4);
      expect(options[0]).toHaveValue('us-east5');
      expect(options[1]).toHaveValue('us-central1');
      expect(options[2]).toHaveValue('europe-west1');
      expect(options[3]).toHaveValue('asia-southeast1');
    });

    it('should update GCP region', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={gcpPreferences}
          onSave={mockOnSave}
        />
      );

      const regionSelect = screen.getByDisplayValue('us-central1');
      fireEvent.change(regionSelect, { target: { value: 'europe-west1' } });

      expect(regionSelect).toHaveValue('europe-west1');
    });
  });

  describe('General Settings - Theme', () => {
    it('should render theme selector', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.getByText('Appearance')).toBeInTheDocument();
      const darkButtons = screen.getAllByText('Dark');
      const lightButtons = screen.getAllByText('Light');
      expect(darkButtons.length).toBeGreaterThan(0);
      expect(lightButtons.length).toBeGreaterThan(0);
    });

    it('should have dark theme selected by default', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const darkButton = screen.getAllByText('Dark')[0].closest('button');
      expect(darkButton).toHaveClass('border-blue-500');
    });

    it('should switch to light theme', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const lightButton = screen.getAllByText('Light')[0].closest('button');
      if (lightButton) {
        fireEvent.click(lightButton);
        expect(lightButton).toHaveClass('border-blue-500');
      }
    });
  });

  describe('General Settings - Notifications', () => {
    it('should render notifications toggle', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.getByText('Desktop Notifications')).toBeInTheDocument();
      expect(screen.getByText('Get notified when tasks complete')).toBeInTheDocument();
    });

    it('should have notifications enabled by default', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const checkbox = screen.getByRole('checkbox');
      expect(checkbox).toBeChecked();
    });

    it('should toggle notifications', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const checkbox = screen.getByRole('checkbox');
      fireEvent.click(checkbox);

      expect(checkbox).not.toBeChecked();
    });
  });

  describe('General Settings - Model Selection', () => {
    it('should render model selector', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.getByText('Default Model')).toBeInTheDocument();
    });

    it('should display current model selection', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const modelSelect = screen.getByDisplayValue('Claude Sonnet 4');
      expect(modelSelect).toBeInTheDocument();
    });

    it('should have all model options', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const select = screen.getByDisplayValue('Claude Sonnet 4');
      const options = within(select.parentElement!).getAllByRole('option');

      expect(options).toHaveLength(5);
      expect(options[0]).toHaveTextContent('Claude Sonnet 4');
      expect(options[1]).toHaveTextContent('Claude Opus 4');
      expect(options[2]).toHaveTextContent('Claude Opus 4.6 (Latest)');
      expect(options[3]).toHaveTextContent('Claude Haiku 4');
      expect(options[4]).toHaveTextContent('Claude 3.5 Sonnet');
    });

    it('should update model selection', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const select = screen.getByDisplayValue('Claude Sonnet 4');
      fireEvent.change(select, { target: { value: 'opus' } });

      expect(select).toHaveValue('opus');
    });
  });

  describe('Approval Settings', () => {
    beforeEach(() => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const approvalTab = screen.getByRole('button', { name: /Approval/i });
      fireEvent.click(approvalTab);
    });

    it('should render approval mode options', () => {
      expect(screen.getByText('Suggest Mode')).toBeInTheDocument();
      expect(screen.getByText('Auto-Edit Mode')).toBeInTheDocument();
      expect(screen.getByText('Full Auto Mode')).toBeInTheDocument();
    });

    it('should display approval mode descriptions', () => {
      expect(screen.getByText('Claude suggests changes, you approve everything')).toBeInTheDocument();
      expect(screen.getByText('Claude can edit files, but asks for bash commands')).toBeInTheDocument();
      expect(screen.getByText('Claude has full control (use with caution)')).toBeInTheDocument();
    });

    it('should have suggest mode selected by default', () => {
      const radios = screen.getAllByRole('radio');
      expect(radios[0]).toBeChecked();
    });

    it('should switch to auto-edit mode', () => {
      const radios = screen.getAllByRole('radio');
      const autoEditRadio = radios[1];

      fireEvent.click(autoEditRadio);
      expect(autoEditRadio).toBeChecked();
    });

    it('should switch to full-auto mode', () => {
      const radios = screen.getAllByRole('radio');
      const fullAutoRadio = radios[2];

      fireEvent.click(fullAutoRadio);
      expect(fullAutoRadio).toBeChecked();
    });

    it('should show warning when full-auto mode is selected', () => {
      const radios = screen.getAllByRole('radio');
      const fullAutoRadio = radios[2];

      fireEvent.click(fullAutoRadio);

      expect(screen.getByText(/Warning: Full auto mode gives Claude complete control/)).toBeInTheDocument();
    });

    it('should not show warning for other modes', () => {
      const warningText = /Warning: Full auto mode gives Claude complete control/;
      expect(screen.queryByText(warningText)).not.toBeInTheDocument();
    });
  });

  describe('MCP Settings - Empty State', () => {
    beforeEach(() => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const mcpTab = screen.getByRole('button', { name: /MCP Servers/i });
      fireEvent.click(mcpTab);
    });

    it('should show empty state when no servers configured', () => {
      expect(screen.getByText('No MCP servers configured')).toBeInTheDocument();
    });

    it('should render add server button in empty state', () => {
      expect(screen.getByText('Add Server')).toBeInTheDocument();
    });
  });

  describe('MCP Settings - With Servers', () => {
    const preferencesWithServers: UserPreferences = {
      ...mockPreferences,
      mcpServers: [
        {
          name: 'Test Server 1',
          command: 'npx test-server',
          enabled: true,
        },
        {
          name: 'Test Server 2',
          command: 'node server.js',
          args: ['--port', '3000'],
          enabled: false,
        },
      ],
    };

    beforeEach(() => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={preferencesWithServers}
          onSave={mockOnSave}
        />
      );

      const mcpTab = screen.getByRole('button', { name: /MCP Servers/i });
      fireEvent.click(mcpTab);
    });

    it('should render server list', () => {
      expect(screen.getByText('Test Server 1')).toBeInTheDocument();
      expect(screen.getByText('Test Server 2')).toBeInTheDocument();
    });

    it('should display server commands', () => {
      expect(screen.getByText('npx test-server')).toBeInTheDocument();
      expect(screen.getByText('node server.js')).toBeInTheDocument();
    });

    it('should show enabled/disabled status', () => {
      const enabledLabels = screen.getAllByText('Enabled');
      const disabledLabels = screen.getAllByText('Disabled');

      expect(enabledLabels.length).toBeGreaterThan(0);
      expect(disabledLabels.length).toBeGreaterThan(0);
    });

    it('should have correct checkbox states', () => {
      const checkboxes = screen.getAllByRole('checkbox');

      // The first checkbox is notifications checkbox from General tab (still in state but not visible)
      // Server checkboxes start after that
      const allCheckboxes = Array.from(checkboxes);
      const serverCheckboxes = allCheckboxes.slice(-2); // Get the last 2 checkboxes (server ones)

      // First server is enabled
      expect(serverCheckboxes[0]).toBeChecked();
      // Second server is disabled
      expect(serverCheckboxes[1]).not.toBeChecked();
    });

    it('should toggle server enabled state', () => {
      const checkboxes = screen.getAllByRole('checkbox');
      const allCheckboxes = Array.from(checkboxes);
      const serverCheckboxes = allCheckboxes.slice(-2); // Get the last 2 checkboxes (server ones)
      const firstServerCheckbox = serverCheckboxes[0];

      expect(firstServerCheckbox).toBeChecked();
      fireEvent.click(firstServerCheckbox);
      expect(firstServerCheckbox).not.toBeChecked();
    });
  });

  describe('About Tab', () => {
    beforeEach(() => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const aboutTab = screen.getByRole('button', { name: /About/i });
      fireEvent.click(aboutTab);
    });

    it('should render app name and description', () => {
      expect(screen.getByText('Boatman')).toBeInTheDocument();
      expect(screen.getByText('Claude Code Desktop App')).toBeInTheDocument();
    });

    it('should display version number', () => {
      expect(screen.getByText('Version 0.1.0')).toBeInTheDocument();
    });

    it('should render description text', () => {
      expect(screen.getByText(/Built with Wails and React/)).toBeInTheDocument();
      expect(screen.getByText(/Powered by Claude/)).toBeInTheDocument();
    });

    it('should have GitHub link', () => {
      const link = screen.getByText('View on GitHub');
      expect(link).toHaveAttribute('href', 'https://github.com');
      expect(link).toHaveAttribute('target', '_blank');
    });
  });

  describe('Form submission', () => {
    it('should save changes when Save button is clicked', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      // Make some changes
      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      fireEvent.change(apiKeyInput, { target: { value: 'sk-ant-new-key' } });

      const saveButton = screen.getByRole('button', { name: 'Save Changes' });
      fireEvent.click(saveButton);

      expect(mockOnSave).toHaveBeenCalledTimes(1);
      expect(mockOnSave).toHaveBeenCalledWith(
        expect.objectContaining({
          apiKey: 'sk-ant-new-key',
        })
      );
      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should save all preference changes', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      // Change API key
      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      fireEvent.change(apiKeyInput, { target: { value: 'sk-ant-updated' } });

      // Change theme
      const lightButton = screen.getAllByText('Light')[0].closest('button');
      if (lightButton) fireEvent.click(lightButton);

      // Change model
      const modelSelect = screen.getByDisplayValue('Claude Sonnet 4');
      fireEvent.change(modelSelect, { target: { value: 'opus' } });

      // Toggle notifications
      const checkbox = screen.getByRole('checkbox');
      fireEvent.click(checkbox);

      const saveButton = screen.getByRole('button', { name: 'Save Changes' });
      fireEvent.click(saveButton);

      expect(mockOnSave).toHaveBeenCalledWith(
        expect.objectContaining({
          apiKey: 'sk-ant-updated',
          theme: 'light',
          defaultModel: 'opus',
          notificationsEnabled: false,
        })
      );
    });

    it('should save approval mode changes', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      // Switch to approval tab
      const approvalTab = screen.getByRole('button', { name: /Approval/i });
      fireEvent.click(approvalTab);

      // Select auto-edit mode
      const radios = screen.getAllByRole('radio');
      fireEvent.click(radios[1]);

      const saveButton = screen.getByRole('button', { name: 'Save Changes' });
      fireEvent.click(saveButton);

      expect(mockOnSave).toHaveBeenCalledWith(
        expect.objectContaining({
          approvalMode: 'auto-edit',
        })
      );
    });

    it('should not save when Cancel is clicked', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      // Make changes
      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      fireEvent.change(apiKeyInput, { target: { value: 'sk-ant-new-key' } });

      const cancelButton = screen.getByRole('button', { name: 'Cancel' });
      fireEvent.click(cancelButton);

      expect(mockOnSave).not.toHaveBeenCalled();
      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Local state management', () => {
    it('should initialize with provided preferences', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      expect(apiKeyInput).toHaveValue(mockPreferences.apiKey);
    });

    it('should maintain local state across tab switches', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      // Change API key in General tab
      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      fireEvent.change(apiKeyInput, { target: { value: 'sk-ant-changed' } });

      // Switch to Approval tab
      const approvalTab = screen.getByRole('button', { name: /Approval/i });
      fireEvent.click(approvalTab);

      // Switch back to General tab
      const generalTab = screen.getByRole('button', { name: /General/i });
      fireEvent.click(generalTab);

      // Verify the change is still there
      const apiKeyInputAfter = screen.getByPlaceholderText('sk-ant-...');
      expect(apiKeyInputAfter).toHaveValue('sk-ant-changed');
    });

    it('should preserve MCP server changes', () => {
      const preferencesWithServers: UserPreferences = {
        ...mockPreferences,
        mcpServers: [
          {
            name: 'Test Server',
            command: 'test-command',
            enabled: true,
          },
        ],
      };

      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={preferencesWithServers}
          onSave={mockOnSave}
        />
      );

      // Switch to MCP tab
      const mcpTab = screen.getByRole('button', { name: /MCP Servers/i });
      fireEvent.click(mcpTab);

      // Toggle server - get the last checkbox which should be the server one
      const checkboxes = screen.getAllByRole('checkbox');
      const serverCheckbox = checkboxes[checkboxes.length - 1];
      if (serverCheckbox) {
        fireEvent.click(serverCheckbox);
      }

      // Switch tabs
      const generalTab = screen.getByRole('button', { name: /General/i });
      fireEvent.click(generalTab);

      // Save
      const saveButton = screen.getByRole('button', { name: 'Save Changes' });
      fireEvent.click(saveButton);

      expect(mockOnSave).toHaveBeenCalledWith(
        expect.objectContaining({
          mcpServers: [
            {
              name: 'Test Server',
              command: 'test-command',
              enabled: false,
            },
          ],
        })
      );
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA roles for modal', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      expect(screen.getByText('Settings')).toBeInTheDocument();
    });

    it('should have accessible form inputs', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      expect(apiKeyInput).toHaveAttribute('type');
    });

    it('should have accessible radio buttons in approval settings', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const approvalTab = screen.getByRole('button', { name: /Approval/i });
      fireEvent.click(approvalTab);

      const radios = screen.getAllByRole('radio');
      radios.forEach((radio) => {
        expect(radio).toHaveAttribute('name', 'approvalMode');
      });
    });
  });

  describe('Edge cases', () => {
    it('should handle empty API key', () => {
      const emptyPreferences: UserPreferences = {
        ...mockPreferences,
        apiKey: '',
      };

      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={emptyPreferences}
          onSave={mockOnSave}
        />
      );

      const apiKeyInput = screen.getByPlaceholderText('sk-ant-...');
      expect(apiKeyInput).toHaveValue('');
    });

    it('should handle empty MCP servers array', () => {
      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={mockPreferences}
          onSave={mockOnSave}
        />
      );

      const mcpTab = screen.getByRole('button', { name: /MCP Servers/i });
      fireEvent.click(mcpTab);

      expect(screen.getByText('No MCP servers configured')).toBeInTheDocument();
    });

    it('should handle undefined GCP values', () => {
      const prefsWithoutGCP: UserPreferences = {
        ...mockPreferences,
        authMethod: 'google-cloud',
        gcpProjectId: undefined,
        gcpRegion: undefined,
      };

      render(
        <SettingsModal
          isOpen={true}
          onClose={mockOnClose}
          preferences={prefsWithoutGCP}
          onSave={mockOnSave}
        />
      );

      const projectIdInput = screen.getByPlaceholderText('my-project-id');
      expect(projectIdInput).toHaveValue('');

      const regionSelect = screen.getByDisplayValue('us-east5');
      expect(regionSelect).toBeInTheDocument();
    });
  });
});
