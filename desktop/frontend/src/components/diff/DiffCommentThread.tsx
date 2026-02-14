import { useState } from 'react';
import { MessageSquare, X, Send } from 'lucide-react';
import { diff } from '../../../wailsjs/go/models';

interface DiffCommentThreadProps {
  comments: diff.DiffComment[];
  lineNum: number;
  hunkId?: string;
  onAddComment: (content: string) => void;
  onDeleteComment: (commentId: string) => void;
}

export function DiffCommentThread({
  comments,
  lineNum,
  hunkId,
  onAddComment,
  onDeleteComment,
}: DiffCommentThreadProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [newComment, setNewComment] = useState('');

  const threadComments = comments.filter(
    (c) => c.lineNum === lineNum && (!hunkId || c.hunkId === hunkId)
  );

  const handleSubmit = () => {
    if (newComment.trim()) {
      onAddComment(newComment.trim());
      setNewComment('');
    }
  };

  return (
    <div className="border-l-2 border-blue-500 bg-slate-900">
      {/* Comment toggle button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 w-full px-4 py-2 text-sm text-slate-300 hover:bg-slate-800 transition-colors"
      >
        <MessageSquare className="w-4 h-4" />
        <span>
          {threadComments.length === 0
            ? 'Add comment'
            : `${threadComments.length} comment${threadComments.length > 1 ? 's' : ''}`}
        </span>
      </button>

      {/* Comments and input */}
      {isOpen && (
        <div className="px-4 py-2 space-y-2">
          {/* Existing comments */}
          {threadComments.map((comment) => (
            <div
              key={comment.id}
              className="bg-slate-800 rounded p-2 relative group"
            >
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1">
                  <div className="text-xs text-slate-400 mb-1">
                    {comment.author || 'You'} •{' '}
                    {new Date(comment.timestamp).toLocaleString()}
                  </div>
                  <div className="text-sm text-slate-200">{comment.content}</div>
                </div>
                <button
                  onClick={() => onDeleteComment(comment.id)}
                  className="opacity-0 group-hover:opacity-100 text-slate-400 hover:text-red-500 transition-opacity"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}

          {/* New comment input */}
          <div className="flex gap-2">
            <input
              type="text"
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
              placeholder="Add a comment..."
              className="flex-1 px-2 py-1 text-sm bg-slate-800 text-slate-200 border border-slate-700 rounded focus:outline-none focus:border-blue-500"
            />
            <button
              onClick={handleSubmit}
              disabled={!newComment.trim()}
              className="px-2 py-1 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <Send className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
