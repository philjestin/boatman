import { useState } from 'react';
import { ChevronLeft, ChevronRight, Sparkles, FolderOpen, Check } from 'lucide-react';
import type { LLMProvider, ProjectLanguage, ScaffoldRequest, ScaffoldResponse } from '../../types';

interface HarnessBuilderProps {
  onScaffold: (req: ScaffoldRequest) => Promise<ScaffoldResponse | null>;
  onSelectFolder: () => Promise<string>;
  onSwitchToRun: () => void;
}

const PROVIDERS: { value: LLMProvider; label: string; description: string }[] = [
  { value: 'claude', label: 'Claude', description: 'Anthropic Claude API' },
  { value: 'openai', label: 'OpenAI', description: 'OpenAI GPT API' },
  { value: 'ollama', label: 'Ollama', description: 'Local models via Ollama' },
  { value: 'generic', label: 'Generic', description: 'Custom LLM provider' },
];

const LANGUAGES: { value: ProjectLanguage; label: string }[] = [
  { value: 'go', label: 'Go' },
  { value: 'typescript', label: 'TypeScript' },
  { value: 'python', label: 'Python' },
  { value: 'ruby', label: 'Ruby' },
  { value: 'generic', label: 'Generic' },
];

export function HarnessBuilder({ onScaffold, onSelectFolder, onSwitchToRun }: HarnessBuilderProps) {
  const [step, setStep] = useState(0);
  const [isGenerating, setIsGenerating] = useState(false);
  const [result, setResult] = useState<ScaffoldResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [projectName, setProjectName] = useState('');
  const [outputDir, setOutputDir] = useState('');
  const [projectLang, setProjectLang] = useState<ProjectLanguage>('go');
  const [provider, setProvider] = useState<LLMProvider>('claude');
  const [includePlanner, setIncludePlanner] = useState(true);
  const [includeTester, setIncludeTester] = useState(true);
  const [includeCostTracking, setIncludeCostTracking] = useState(false);
  const [maxIterations, setMaxIterations] = useState(3);
  const [reviewCriteria, setReviewCriteria] = useState('');

  const steps = ['Project', 'Provider', 'Options', 'Review'];

  const canNext = () => {
    if (step === 0) return projectName.trim() !== '';
    return true;
  };

  const handleBrowse = async () => {
    const folder = await onSelectFolder();
    if (folder) setOutputDir(folder);
  };

  const handleGenerate = async () => {
    setIsGenerating(true);
    setError(null);

    try {
      const res = await onScaffold({
        projectName: projectName.trim(),
        outputDir,
        provider,
        projectLang,
        includePlanner,
        includeTester,
        includeCostTracking,
        maxIterations,
        reviewCriteria,
      });
      setResult(res);
    } catch (err) {
      setError(String(err));
    } finally {
      setIsGenerating(false);
    }
  };

  const renderStep = () => {
    switch (step) {
      case 0:
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Project Name</label>
              <input
                type="text"
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
                placeholder="e.g., my-agent"
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                autoFocus
              />
              <p className="text-xs text-slate-500 mt-1">Go module path for the generated project</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Output Directory</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={outputDir}
                  onChange={(e) => setOutputDir(e.target.value)}
                  placeholder="~/.boatman/harnesses/{name}/ (default)"
                  className="flex-1 px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm"
                />
                <button
                  onClick={handleBrowse}
                  className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-slate-300 rounded-md transition-colors text-sm"
                >
                  <FolderOpen className="w-4 h-4" />
                </button>
              </div>
              <p className="text-xs text-slate-500 mt-1">Leave empty for default location</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Project Language</label>
              <div className="grid grid-cols-3 gap-2">
                {LANGUAGES.map((lang) => (
                  <button
                    key={lang.value}
                    onClick={() => setProjectLang(lang.value)}
                    className={`px-3 py-2 text-sm rounded-md border transition-colors ${
                      projectLang === lang.value
                        ? 'border-blue-500 bg-blue-500/20 text-blue-300'
                        : 'border-slate-700 bg-slate-800/50 text-slate-400 hover:border-slate-600'
                    }`}
                  >
                    {lang.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        );

      case 1:
        return (
          <div className="space-y-4">
            <label className="block text-sm font-medium text-slate-300 mb-2">LLM Provider</label>
            <div className="grid grid-cols-2 gap-3">
              {PROVIDERS.map((p) => (
                <button
                  key={p.value}
                  onClick={() => setProvider(p.value)}
                  className={`p-4 text-left rounded-lg border transition-colors ${
                    provider === p.value
                      ? 'border-blue-500 bg-blue-500/20'
                      : 'border-slate-700 bg-slate-800/50 hover:border-slate-600'
                  }`}
                >
                  <div className="text-sm font-medium text-slate-200">{p.label}</div>
                  <div className="text-xs text-slate-400 mt-1">{p.description}</div>
                </button>
              ))}
            </div>
          </div>
        );

      case 2:
        return (
          <div className="space-y-4">
            <div className="space-y-3">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={includePlanner}
                  onChange={(e) => setIncludePlanner(e.target.checked)}
                  className="w-4 h-4 rounded border-slate-600 bg-slate-800 text-blue-500 focus:ring-blue-500"
                />
                <div>
                  <div className="text-sm text-slate-200">Include Planner</div>
                  <div className="text-xs text-slate-500">Generate a planning phase before implementation</div>
                </div>
              </label>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={includeTester}
                  onChange={(e) => setIncludeTester(e.target.checked)}
                  className="w-4 h-4 rounded border-slate-600 bg-slate-800 text-blue-500 focus:ring-blue-500"
                />
                <div>
                  <div className="text-sm text-slate-200">Include Tester</div>
                  <div className="text-xs text-slate-500">Run tests after each implementation iteration</div>
                </div>
              </label>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={includeCostTracking}
                  onChange={(e) => setIncludeCostTracking(e.target.checked)}
                  className="w-4 h-4 rounded border-slate-600 bg-slate-800 text-blue-500 focus:ring-blue-500"
                />
                <div>
                  <div className="text-sm text-slate-200">Include Cost Tracking</div>
                  <div className="text-xs text-slate-500">Track token usage and estimated costs</div>
                </div>
              </label>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Max Iterations</label>
              <input
                type="number"
                value={maxIterations}
                onChange={(e) => setMaxIterations(parseInt(e.target.value) || 1)}
                min={1}
                max={20}
                className="w-24 px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Review Criteria (optional)</label>
              <textarea
                value={reviewCriteria}
                onChange={(e) => setReviewCriteria(e.target.value)}
                placeholder="e.g., Focus on error handling and test coverage"
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                rows={3}
              />
            </div>
          </div>
        );

      case 3:
        if (result) {
          return (
            <div className="space-y-4">
              <div className="flex items-center gap-2 text-green-400">
                <Check className="w-5 h-5" />
                <span className="text-sm font-medium">Project generated successfully</span>
              </div>
              <div>
                <p className="text-xs text-slate-400 mb-1">Output Directory</p>
                <p className="text-sm font-mono text-slate-200">{result.outputDir}</p>
              </div>
              <div>
                <p className="text-xs text-slate-400 mb-2">Files Created</p>
                <div className="space-y-1">
                  {result.filesCreated.map((file) => (
                    <p key={file} className="text-sm font-mono text-slate-300">{file}</p>
                  ))}
                </div>
              </div>
              <button
                onClick={onSwitchToRun}
                className="w-full px-4 py-2 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors"
              >
                Open in Run Mode
              </button>
            </div>
          );
        }

        return (
          <div className="space-y-4">
            <h3 className="text-sm font-medium text-slate-300">Configuration Summary</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Project Name</span>
                <span className="text-slate-200 font-mono">{projectName}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Output</span>
                <span className="text-slate-200 font-mono text-xs">{outputDir || '(default)'}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Language</span>
                <span className="text-slate-200">{projectLang}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Provider</span>
                <span className="text-slate-200">{provider}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Planner</span>
                <span className="text-slate-200">{includePlanner ? 'Yes' : 'No'}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Tester</span>
                <span className="text-slate-200">{includeTester ? 'Yes' : 'No'}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Cost Tracking</span>
                <span className="text-slate-200">{includeCostTracking ? 'Yes' : 'No'}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Max Iterations</span>
                <span className="text-slate-200">{maxIterations}</span>
              </div>
              {reviewCriteria && (
                <div className="flex justify-between py-1.5 border-b border-slate-800">
                  <span className="text-slate-400">Review Criteria</span>
                  <span className="text-slate-200 text-xs max-w-[200px] text-right">{reviewCriteria}</span>
                </div>
              )}
            </div>

            {error && (
              <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-md text-sm text-red-400">
                {error}
              </div>
            )}

            <button
              onClick={handleGenerate}
              disabled={isGenerating}
              className="w-full flex items-center justify-center gap-2 px-4 py-2.5 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Sparkles className="w-4 h-4" />
              {isGenerating ? 'Generating...' : 'Generate Project'}
            </button>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Progress bar */}
      <div className="px-6 py-4 border-b border-slate-700">
        <div className="flex items-center gap-2">
          {steps.map((s, i) => (
            <div key={s} className="flex items-center gap-2">
              <div
                className={`flex items-center justify-center w-7 h-7 rounded-full text-xs font-medium ${
                  i < step
                    ? 'bg-blue-500 text-white'
                    : i === step
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500'
                    : 'bg-slate-800 text-slate-500 border border-slate-700'
                }`}
              >
                {i < step ? <Check className="w-3.5 h-3.5" /> : i + 1}
              </div>
              <span className={`text-xs ${i === step ? 'text-slate-200' : 'text-slate-500'}`}>{s}</span>
              {i < steps.length - 1 && <div className="w-8 h-px bg-slate-700" />}
            </div>
          ))}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6">
        {renderStep()}
      </div>

      {/* Navigation */}
      {!result && (
        <div className="flex items-center justify-between px-6 py-4 border-t border-slate-700">
          <button
            onClick={() => setStep((s) => Math.max(0, s - 1))}
            disabled={step === 0}
            className="flex items-center gap-1 px-3 py-2 text-sm text-slate-400 hover:text-slate-200 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            <ChevronLeft className="w-4 h-4" />
            Back
          </button>
          {step < steps.length - 1 && (
            <button
              onClick={() => setStep((s) => Math.min(steps.length - 1, s + 1))}
              disabled={!canNext()}
              className="flex items-center gap-1 px-4 py-2 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Next
              <ChevronRight className="w-4 h-4" />
            </button>
          )}
        </div>
      )}
    </div>
  );
}
