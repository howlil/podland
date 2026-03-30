// Analytics tracking utility for Podland
// Integrates with your analytics service (e.g., Plausible, Google Analytics, etc.)

type EventType = 
  | 'page_view'
  | 'vm_create'
  | 'vm_start'
  | 'vm_stop'
  | 'vm_restart'
  | 'vm_delete'
  | 'vm_pin'
  | 'vm_unpin'
  | 'user_login'
  | 'user_logout'
  | 'admin_action';

interface AnalyticsEvent {
  event: EventType;
  properties?: Record<string, any>;
  timestamp?: number;
}

/**
 * Track analytics event
 * Replace with your actual analytics service
 */
export function trackEvent({ event, properties, timestamp }: AnalyticsEvent) {
  // In production, replace with actual analytics service
  // Example: plausible(event, { props: properties });
  // Example: gtag('event', event, properties);
  
  // For now, log to console in development
  if (process.env.NODE_ENV === 'development') {
    console.log('[Analytics]', event, properties);
  }
  
  // Send to analytics endpoint
  try {
    fetch('/api/analytics', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        event,
        properties,
        timestamp: timestamp || Date.now(),
        url: window.location.href,
        referrer: document.referrer,
      }),
    }).catch(() => {
      // Silently fail - analytics shouldn't break the app
    });
  } catch {
    // Silently fail
  }
}

/**
 * Track page view
 */
export function trackPageView(path?: string) {
  trackEvent({
    event: 'page_view',
    properties: {
      path: path || window.location.pathname,
      title: document.title,
    },
  });
}

/**
 * Track VM actions
 */
export const trackVM = {
  create: (tier: string, os: string) =>
    trackEvent({ event: 'vm_create', properties: { tier, os } }),
  
  start: (vmId: string, vmName: string) =>
    trackEvent({ event: 'vm_start', properties: { vmId, vmName } }),
  
  stop: (vmId: string, vmName: string) =>
    trackEvent({ event: 'vm_stop', properties: { vmId, vmName } }),
  
  restart: (vmId: string, vmName: string) =>
    trackEvent({ event: 'vm_restart', properties: { vmId, vmName } }),
  
  delete: (vmId: string, vmName: string) =>
    trackEvent({ event: 'vm_delete', properties: { vmId, vmName } }),
  
  pin: (vmId: string, vmName: string) =>
    trackEvent({ event: 'vm_pin', properties: { vmId, vmName } }),
  
  unpin: (vmId: string, vmName: string) =>
    trackEvent({ event: 'vm_unpin', properties: { vmId, vmName } }),
};

/**
 * Track admin actions
 */
export const trackAdmin = {
  action: (action: string, target: string, details?: Record<string, any>) =>
    trackEvent({ event: 'admin_action', properties: { action, target, ...details } }),
};

/**
 * Initialize analytics on app start
 */
export function initAnalytics() {
  // Track initial page view
  trackPageView();
  
  // Track subsequent navigation
  const originalPushState = history.pushState;
  history.pushState = function(...args) {
    originalPushState.apply(this, args);
    trackPageView();
  };
  
  const originalReplaceState = history.replaceState;
  history.replaceState = function(...args) {
    originalReplaceState.apply(this, args);
    trackPageView();
  };
  
  window.addEventListener('popstate', () => {
    trackPageView();
  });
}
