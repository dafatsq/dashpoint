// React 18+ enables the `act` environment flag via a global that React expects
// application code to declare itself (it ships without type definitions).
// Test suites set `globalThis.IS_REACT_ACT_ENVIRONMENT = true` around renders.
declare global {
  var IS_REACT_ACT_ENVIRONMENT: boolean | undefined;
}

export {};
