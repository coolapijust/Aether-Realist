// Icon names for the application
// Using Lucide React icons

export type IconName = 
  | 'home'
  | 'globe'
  | 'shield'
  | 'activity'
  | 'file-text'
  | 'settings'
  | 'power'
  | 'rotate-cw'
  | 'plus'
  | 'minus'
  | 'x'
  | 'check'
  | 'chevron-down'
  | 'chevron-up'
  | 'moon'
  | 'sun'
  | 'wifi'
  | 'wifi-off'
  | 'server'
  | 'layers'
  | 'bar-chart-2'
  | 'clock'
  | 'alert-circle'
  | 'info'
  | 'trash-2'
  | 'refresh-cw'
  | 'play'
  | 'pause'
  | 'save';

// Icon mapping for sidebar navigation
export const navigationIcons = {
  dashboard: 'home',
  proxy: 'globe',
  rules: 'shield',
  connections: 'activity',
  logs: 'file-text',
  settings: 'settings',
} as const;
