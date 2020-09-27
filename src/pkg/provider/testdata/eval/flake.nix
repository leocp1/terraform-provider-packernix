{
  description = "A flake to test packernix_eval";

  outputs = { ... }: rec {
    fib = (import ./fib.nix);
    fib10 = (fib { x = 10; }).out;
  };
}
