export interface Scope {
  org_id: string;
  team_id: string;
  repo_id: string;
}

export interface Run {
  id: string;
  scope: Scope;
  user_id: string;
  status: string;
  prompt: string;
  total_cost_usd: number;
  iterations: number;
  files_changed: string[];
  duration: number;
  created_at: string;
  updated_at: string;
}

export interface Pattern {
  id: string;
  scope: Scope;
  type: string;
  description: string;
  example?: string;
  file_matcher?: string;
  weight: number;
  usage_count: number;
  success_rate: number;
  created_at: string;
  updated_at: string;
}

export interface Policy {
  id: string;
  scope: Scope;
  max_iterations?: number;
  max_cost_per_run?: number;
  max_files_changed?: number;
  allowed_models?: string[];
  blocked_patterns?: string[];
  require_tests: boolean;
  require_review: boolean;
  updated_at: string;
}

export interface UsageSummary {
  period: string;
  total_runs: number;
  total_cost_usd: number;
  input_tokens: number;
  output_tokens: number;
}

export interface Budget {
  id: string;
  scope: Scope;
  monthly_limit: number;
  daily_limit: number;
  per_run_limit: number;
  alert_at: number;
  updated_at: string;
}

export interface Event {
  id: string;
  run_id?: string;
  scope: Scope;
  type: string;
  name?: string;
  message?: string;
  data?: Record<string, unknown>;
  version: number;
  created_at: string;
}
