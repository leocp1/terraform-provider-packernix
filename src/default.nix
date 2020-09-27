{ stdenv
, lib
, buildGoModule
, gitignoreSource
, nixFilter
, terraform
, nixFlakes
, nix
}:
let
  nix = nixFlakes;
in buildGoModule rec {
  pname = "terraform-provider-packernix";
  version = "0.0.1";
  src = gitignoreSource ./.;
  vendorSha256 = "sha256-2grMalJN+HA3li/EwWTCD2//XvNZMSpn9fLfdtIpXy4=";
  checkInputs = [ terraform ];
  nativeBuildInputs = [ nix ];
  patchPhase = ''
    substituteInPlace ./pkg/patches/patches.go \
      --subst-var-by nix "${nix}" \
      --subst-var-by out "$out"
  '';
  preCheck = ''
    patchShebangs --build ./pkg/provider/testdata/external/
  '';
  doCheck = true;
  meta = with stdenv.lib; {
    description = "Terraform provider for nix provisioned packer images";
    license = licenses.mpl20;
    platforms = platforms.linux;
  };
}
