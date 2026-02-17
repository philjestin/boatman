import { useState, useRef, useEffect } from 'react';
import { ChevronUp, Check, type LucideIcon } from 'lucide-react';

interface PillDropdownOption {
  value: string;
  label: string;
}

interface PillDropdownProps {
  options: readonly PillDropdownOption[];
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  icon?: LucideIcon;
}

export function PillDropdown({ options, value, onChange, disabled = false, icon: Icon }: PillDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const selectedOption = options.find((o) => o.value === value);
  const displayLabel = selectedOption?.label ?? value;

  // Close on click outside
  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  return (
    <div ref={containerRef} className="relative">
      {/* Pill button */}
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={`flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-full border transition-colors ${
          disabled
            ? 'border-slate-700 bg-slate-800 text-slate-500 cursor-not-allowed'
            : isOpen
            ? 'border-blue-500 bg-slate-700 text-slate-200'
            : 'border-slate-600 bg-slate-800 text-slate-300 hover:border-slate-500 hover:text-slate-200'
        }`}
      >
        {Icon && <Icon className="w-3 h-3" />}
        <span>{displayLabel}</span>
        <ChevronUp className={`w-3 h-3 transition-transform ${isOpen ? '' : 'rotate-180'}`} />
      </button>

      {/* Dropdown menu (opens upward) */}
      {isOpen && (
        <div className="absolute bottom-full left-0 mb-1 min-w-[160px] py-1 bg-slate-800 border border-slate-600 rounded-lg shadow-xl z-50">
          {options.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => {
                onChange(option.value);
                setIsOpen(false);
              }}
              className={`flex items-center justify-between w-full px-3 py-1.5 text-xs text-left transition-colors ${
                option.value === value
                  ? 'text-blue-400 bg-slate-700/50'
                  : 'text-slate-300 hover:bg-slate-700 hover:text-slate-200'
              }`}
            >
              <span>{option.label}</span>
              {option.value === value && <Check className="w-3 h-3" />}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
