{ pkgs ? import ../../../nix { } }:
let aizeld = (pkgs.callPackage ../../../. { });
in
aizeld.overrideAttrs (oldAttrs: {
  patches = oldAttrs.patches or [ ] ++ [
    ./broken-aizeld.patch
  ];
})
