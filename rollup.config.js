import typescript from "@rollup/plugin-typescript";
// For resolving node_modules
import nodeResolve from "@rollup/plugin-node-resolve";

export default {
  input: "ts/index.ts",
  output: {
    dir: "static/dist",
    format: "es",
  },
  plugins: [typescript(), nodeResolve()],
};
