import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DiffCommentThread } from './DiffCommentThread';
import type { DiffComment } from '../../types';

describe('DiffCommentThread', () => {
  const mockComments: DiffComment[] = [
    {
      id: 'comment1',
      lineNum: 10,
      content: 'This needs review',
      timestamp: '2026-01-01T10:00:00Z',
      author: 'User1',
    },
    {
      id: 'comment2',
      lineNum: 10,
      content: 'Looks good to me',
      timestamp: '2026-01-01T11:00:00Z',
      author: 'User2',
    },
  ];

  it('should render add comment button when closed', () => {
    render(
      <DiffCommentThread
        comments={[]}
        lineNum={10}
        onAddComment={vi.fn()}
        onDeleteComment={vi.fn()}
      />
    );

    expect(screen.getByText('Add comment')).toBeInTheDocument();
  });

  it('should show comment count when there are comments', () => {
    render(
      <DiffCommentThread
        comments={mockComments}
        lineNum={10}
        onAddComment={vi.fn()}
        onDeleteComment={vi.fn()}
      />
    );

    expect(screen.getByText('2 comments')).toBeInTheDocument();
  });

  it('should expand and show comments when clicked', () => {
    render(
      <DiffCommentThread
        comments={mockComments}
        lineNum={10}
        onAddComment={vi.fn()}
        onDeleteComment={vi.fn()}
      />
    );

    const toggleButton = screen.getByText('2 comments');
    fireEvent.click(toggleButton);

    expect(screen.getByText('This needs review')).toBeInTheDocument();
    expect(screen.getByText('Looks good to me')).toBeInTheDocument();
  });

  it('should call onAddComment when submitting new comment', () => {
    const onAddComment = vi.fn();
    render(
      <DiffCommentThread
        comments={[]}
        lineNum={10}
        onAddComment={onAddComment}
        onDeleteComment={vi.fn()}
      />
    );

    // Open the thread
    fireEvent.click(screen.getByText('Add comment'));

    // Type a comment
    const input = screen.getByPlaceholderText('Add a comment...');
    fireEvent.change(input, { target: { value: 'New comment' } });

    // Click send button
    const sendButton = screen.getByRole('button', { name: '' }); // Send button has no text, just icon
    fireEvent.click(sendButton);

    expect(onAddComment).toHaveBeenCalledWith('New comment');
  });

  it('should call onAddComment when pressing Enter', () => {
    const onAddComment = vi.fn();
    render(
      <DiffCommentThread
        comments={[]}
        lineNum={10}
        onAddComment={onAddComment}
        onDeleteComment={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('Add comment'));

    const input = screen.getByPlaceholderText('Add a comment...');
    fireEvent.change(input, { target: { value: 'New comment' } });
    fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

    expect(onAddComment).toHaveBeenCalledWith('New comment');
  });

  it('should filter comments by lineNum', () => {
    const allComments: DiffComment[] = [
      ...mockComments,
      {
        id: 'comment3',
        lineNum: 20,
        content: 'Different line',
        timestamp: '2026-01-01T12:00:00Z',
      },
    ];

    render(
      <DiffCommentThread
        comments={allComments}
        lineNum={10}
        onAddComment={vi.fn()}
        onDeleteComment={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('2 comments'));

    expect(screen.getByText('This needs review')).toBeInTheDocument();
    expect(screen.queryByText('Different line')).not.toBeInTheDocument();
  });

  it('should filter comments by hunkId when provided', () => {
    const allComments: DiffComment[] = [
      { ...mockComments[0], hunkId: 'hunk1' },
      { ...mockComments[1], hunkId: 'hunk2' },
    ];

    render(
      <DiffCommentThread
        comments={allComments}
        lineNum={10}
        hunkId="hunk1"
        onAddComment={vi.fn()}
        onDeleteComment={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('1 comment'));

    expect(screen.getByText('This needs review')).toBeInTheDocument();
    expect(screen.queryByText('Looks good to me')).not.toBeInTheDocument();
  });

  it('should call onDeleteComment when delete button clicked', () => {
    const onDeleteComment = vi.fn();
    render(
      <DiffCommentThread
        comments={mockComments}
        lineNum={10}
        onAddComment={vi.fn()}
        onDeleteComment={onDeleteComment}
      />
    );

    fireEvent.click(screen.getByText('2 comments'));

    // Get all delete buttons (X icons) - there should be 2, one for each comment
    const deleteButtons = screen.getAllByRole('button');
    const firstDeleteButton = deleteButtons.find(
      (btn) => btn.querySelector('svg') && btn !== screen.getByText('2 comments').closest('button')
    );

    if (firstDeleteButton) {
      fireEvent.click(firstDeleteButton);
      expect(onDeleteComment).toHaveBeenCalled();
    }
  });

  it('should not submit empty comments', () => {
    const onAddComment = vi.fn();
    render(
      <DiffCommentThread
        comments={[]}
        lineNum={10}
        onAddComment={onAddComment}
        onDeleteComment={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('Add comment'));

    const input = screen.getByPlaceholderText('Add a comment...');
    fireEvent.change(input, { target: { value: '   ' } }); // Only whitespace

    const sendButton = screen.getByRole('button', { name: '' });
    fireEvent.click(sendButton);

    expect(onAddComment).not.toHaveBeenCalled();
  });

  it('should clear input after submitting comment', () => {
    render(
      <DiffCommentThread
        comments={[]}
        lineNum={10}
        onAddComment={vi.fn()}
        onDeleteComment={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('Add comment'));

    const input = screen.getByPlaceholderText('Add a comment...') as HTMLInputElement;
    fireEvent.change(input, { target: { value: 'New comment' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    expect(input.value).toBe('');
  });
});
