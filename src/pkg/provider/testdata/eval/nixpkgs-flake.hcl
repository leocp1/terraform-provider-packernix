provider packernix {}

data "packernix_eval" "nixpkgs" {
  installable = "nixpkgs#path"
  out_link = "{{.TempDir}}/nixpkgs-eval-flake"
}
