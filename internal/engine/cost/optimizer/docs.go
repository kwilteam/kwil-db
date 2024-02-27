/*
Package optimizer provides xxx.

LogicalOptimizeRules:
  - projection pushdown
  - selection pushdown
  - insert additional projection to overcome the side effect of pushdowns; e.g.
    from select b from t where a = 1;
    to select b from t[a,b] where a = 1;
    we add extra projection `a` to the scan
  - combine selection and corss product into join
*/
package optimizer
