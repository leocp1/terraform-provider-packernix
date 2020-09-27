provider packernix {}

data "packernix_build" "installable" {
  installable = "nixpkgs#hello"
  clear_env = true
}
