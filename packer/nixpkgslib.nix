# Functions from nixpkgs.lib
let
  inherit (builtins)
    head tail length concatMap
    mapAttrs attrNames listToAttrs catAttrs isAttrs
    ;
in
rec {
  /* Utility function that creates a {name, value} pair as expected by
     builtins.listToAttrs.

     Example:
       nameValuePair "some" 6
       => { name = "some"; value = 6; }
  */
  nameValuePair = name: value: { inherit name value; };

  /*
     Filter an attribute set by removing all attributes for which the
     given predicate return false.

     Example:
       filterAttrs (n: v: n == "foo") { foo = 1; bar = 2; }
       => { foo = 1; }
  */
  filterAttrs = pred: set:
    listToAttrs (concatMap (name: let v = set.${name}; in if pred name v then [ (nameValuePair name v) ] else []) (attrNames set));

  /* Merge sets of attributes and use the function f to merge attributes
     values.

     Example:
       zipAttrsWithNames ["a"] (name: vs: vs) [{a = "x";} {a = "y"; b = "z";}]
       => { a = ["x" "y"]; }
  */
  zipAttrsWithNames = names: f: sets:
    listToAttrs (
      map (
        name: {
          inherit name;
          value = f name (catAttrs name sets);
        }
      ) names
    );

  /* Implementation note: Common names  appear multiple times in the list of
     names, hopefully this does not affect the system because the maximal
     laziness avoid computing twice the same expression and listToAttrs does
     not care about duplicated attribute names.

     Example:
       zipAttrsWith (name: values: values) [{a = "x";} {a = "y"; b = "z";}]
       => { a = ["x" "y"]; b = ["z"] }
  */
  zipAttrsWith = f: sets: zipAttrsWithNames (concatMap attrNames sets) f sets;
  /* Like `zipAttrsWith' with `(name: values: values)' as the function.

    Example:
      zipAttrs [{a = "x";} {a = "y"; b = "z";}]
      => { a = ["x" "y"]; b = ["z"] }
  */
  zipAttrs = zipAttrsWith (name: values: values);

  /* Does the same as the update operator '//' except that attributes are
     merged until the given predicate is verified.  The predicate should
     accept 3 arguments which are the path to reach the attribute, a part of
     the first attribute set and a part of the second attribute set.  When
     the predicate is verified, the value of the first attribute set is
     replaced by the value of the second attribute set.

     Example:
       recursiveUpdateUntil (path: l: r: path == ["foo"]) {
         # first attribute set
         foo.bar = 1;
         foo.baz = 2;
         bar = 3;
       } {
         #second attribute set
         foo.bar = 1;
         foo.quz = 2;
         baz = 4;
       }

       returns: {
         foo.bar = 1; # 'foo.*' from the second set
         foo.quz = 2; #
         bar = 3;     # 'bar' from the first set
         baz = 4;     # 'baz' from the second set
       }

     */
  recursiveUpdateUntil = pred: lhs: rhs:
    let
      f = attrPath:
        zipAttrsWith (
          n: values:
            let
              here = attrPath ++ [ n ];
            in
              if tail values == []
              || pred here (head (tail values)) (head values) then
                head values
              else
                f here values
        );
    in
      f [] [ rhs lhs ];

  /* A recursive variant of the update operator ‘//’.  The recursion
     stops when one of the attribute values is not an attribute set,
     in which case the right hand side value takes precedence over the
     left hand side value.

     Example:
       recursiveUpdate {
         boot.loader.grub.enable = true;
         boot.loader.grub.device = "/dev/hda";
       } {
         boot.loader.grub.device = "";
       }

       returns: {
         boot.loader.grub.enable = true;
         boot.loader.grub.device = "";
       }

     */
  recursiveUpdate = lhs: rhs:
    recursiveUpdateUntil (
      path: lhs: rhs:
        !(isAttrs lhs && isAttrs rhs)
    ) lhs rhs;
}
