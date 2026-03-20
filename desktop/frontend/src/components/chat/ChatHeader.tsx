import { Terminal, FolderOpen, Bot, Zap, Hash } from 'lucide-react';

interface ChatHeaderProps {
  projectPath: string;
  model?: string;
  reasoningEffort?: string;
  mode?: string;
  messageCount: number;
}

export function ChatHeader({ projectPath, model, reasoningEffort, mode, messageCount }: ChatHeaderProps) {
  const projectName = projectPath.split('/').pop() || projectPath;
  const isClaudeChat = !mode || mode === 'standard' || mode === '';

  if (!isClaudeChat) return null;

  return (
    <div className="flex items-center gap-3 px-4 py-2 bg-slate-800/80 border-b border-slate-700/50 text-xs text-slate-400">
      <div className="flex items-center gap-1.5">
        <Terminal className="w-3.5 h-3.5 text-blue-400" />
        <span className="text-slate-200 font-medium">Claude</span>
      </div>
      <span className="text-slate-600">|</span>
      <div className="flex items-center gap-1">
        <FolderOpen className="w-3 h-3" />
        <span className="truncate max-w-[200px]" title={projectPath}>{projectName}</span>
      </div>
      {model && (
        <>
          <span className="text-slate-600">|</span>
          <div className="flex items-center gap-1">
            <Bot className="w-3 h-3" />
            <span>{model}</span>
          </div>
        </>
      )}
      {reasoningEffort && (
        <>
          <span className="text-slate-600">|</span>
          <div className="flex items-center gap-1">
            <Zap className="w-3 h-3" />
            <span>{reasoningEffort}</span>
          </div>
        </>
      )}
      <div className="ml-auto flex items-center gap-1">
        <Hash className="w-3 h-3" />
        <span>{messageCount} messages</span>
      </div>
    </div>
  );
}
