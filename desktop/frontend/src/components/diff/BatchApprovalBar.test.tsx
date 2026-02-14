import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BatchApprovalBar } from './BatchApprovalBar';

describe('BatchApprovalBar', () => {
  it('should not render when no files are selected', () => {
    const { container } = render(
      <BatchApprovalBar
        selectedCount={0}
        totalCount={10}
        onApproveSelected={vi.fn()}
        onRejectSelected={vi.fn()}
        onClearSelection={vi.fn()}
      />
    );

    expect(container.firstChild).toBeNull();
  });

  it('should render when files are selected', () => {
    render(
      <BatchApprovalBar
        selectedCount={3}
        totalCount={10}
        onApproveSelected={vi.fn()}
        onRejectSelected={vi.fn()}
        onClearSelection={vi.fn()}
      />
    );

    expect(screen.getByText(/3/)).toBeInTheDocument();
    expect(screen.getByText(/10/)).toBeInTheDocument();
    expect(screen.getByText(/selected/)).toBeInTheDocument();
  });

  it('should show singular "file" for 1 total file', () => {
    const { container } = render(
      <BatchApprovalBar
        selectedCount={1}
        totalCount={1}
        onApproveSelected={vi.fn()}
        onRejectSelected={vi.fn()}
        onClearSelection={vi.fn()}
      />
    );

    const text = container.textContent;
    expect(text).toContain('1');
    expect(text).toContain('file selected');
    expect(text).not.toContain('files');
  });

  it('should call onApproveSelected when approve button clicked', () => {
    const onApproveSelected = vi.fn();
    render(
      <BatchApprovalBar
        selectedCount={3}
        totalCount={10}
        onApproveSelected={onApproveSelected}
        onRejectSelected={vi.fn()}
        onClearSelection={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('Approve Selected'));
    expect(onApproveSelected).toHaveBeenCalledOnce();
  });

  it('should call onRejectSelected when reject button clicked', () => {
    const onRejectSelected = vi.fn();
    render(
      <BatchApprovalBar
        selectedCount={3}
        totalCount={10}
        onApproveSelected={vi.fn()}
        onRejectSelected={onRejectSelected}
        onClearSelection={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('Reject Selected'));
    expect(onRejectSelected).toHaveBeenCalledOnce();
  });

  it('should call onClearSelection when clear button clicked', () => {
    const onClearSelection = vi.fn();
    render(
      <BatchApprovalBar
        selectedCount={3}
        totalCount={10}
        onApproveSelected={vi.fn()}
        onRejectSelected={vi.fn()}
        onClearSelection={onClearSelection}
      />
    );

    fireEvent.click(screen.getByText('Clear'));
    expect(onClearSelection).toHaveBeenCalledOnce();
  });

  it('should display all action buttons', () => {
    render(
      <BatchApprovalBar
        selectedCount={5}
        totalCount={10}
        onApproveSelected={vi.fn()}
        onRejectSelected={vi.fn()}
        onClearSelection={vi.fn()}
      />
    );

    expect(screen.getByText('Clear')).toBeInTheDocument();
    expect(screen.getByText('Reject Selected')).toBeInTheDocument();
    expect(screen.getByText('Approve Selected')).toBeInTheDocument();
  });

  it('should have fixed positioning at bottom', () => {
    const { container } = render(
      <BatchApprovalBar
        selectedCount={1}
        totalCount={10}
        onApproveSelected={vi.fn()}
        onRejectSelected={vi.fn()}
        onClearSelection={vi.fn()}
      />
    );

    const bar = container.firstChild as HTMLElement;
    expect(bar.className).toContain('fixed');
    expect(bar.className).toContain('bottom-0');
  });
});
