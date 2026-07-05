export const environment = {
  apiUrl: '',
  // LINE Login (EL-111): shows the "Log in with LINE" button. On in dev so the
  // flow is exercised locally + in e2e. Prod flips on once the LINE Login channel
  // is live (EL-112).
  lineLoginEnabled: true,
};
