# Path conflict validation in APIRule v2alpha1

APIRule v2alpha1 introduces support for wildcard paths in the `path` field of the `APIRule` CRD. 
This allows you to define a single `APIRule` that matches multiple request paths.
However, this also introduces the possibility of path conflicts. 
A path conflict occurs when two or more `APIRule` resources match the same path.

To prevent path conflicts, the API Gateway Operator validates the paths of all `APIRule` resources in the cluster.
To make sure that the structure supports wildcard paths, the operator uses an algorithm implemented with a modified version of the `trie` data structure,
defined in the `internal/path/segment_trie.go` file.

Stored paths can contain the `{*}` and `{**}` operators:
  - operator `{*}` is used to match a single segment in a path, and may include a prefix and/or suffix.
  - operator `{**}` is used any number of segments in a path, and may include a prefix and/or suffix.
    It must be the last operator in the stored path (this is not validated but is assumed to be true).

It does two things that are not done by a regular trie:
  - Nodes that are pointed to by `{**}` don't store their children,
    instead they store the path suffix that exists after `{**}`.
    Possible paths are found by comparing the suffixes of paths with the same segments that precede `{**}`.
    In case there is no suffix after `{**}`, the node stores `""` as the suffix.
    New paths starting with the same pattern before `{**}` are stored in the same node, and the suffix list is updated.
    This can be done, since the `{**}` operator must be the last operator in any path.
  - `{*}` nodes are stored just like any other exact segment node, but are always included in path search.

An example of the data structure generated from paths:
- `"c/a"`
- `"b/ar"`
- `"b/{*}"`
- `"b/{*}/c"`
- `"b/{**}/a/b"`
- `"b/{**}/a/d"`
- `"d/{**}"`

can be seen in the following diagram:

[![Path trie](../assets/validation-trie.svg)](../assets/validation-trie.svg)

## Validation process

The validation process is performed during the creation of the `trie` data structure.
Before any `path` is added to the `trie`, the operator checks if an existing path in the `trie` is a prefix of the new path.

If no conflict is found, the new path is added to the `trie` and the process continues with the next path.

### Exact path validation

For the most simple case where a conflict is checked between the paths already stored in the `trie` and a new exact path, 
the path search algorithm takes into account the following possible cases:
1. The current search node is a `literal` node - not an operator.
2. The current search node is a `{*}` node.
3. The current search node is a `{**}` node.

In the first case, the algorithm checks if the current segment is equal to the currently checked segment in the path.
If so, the search continues recursively with the next segment in the path.

In the second case, the algorithm does not need to check the current segment, as the `{*}` operator matches any segment.
The search continues recursively with the next segment in the path.

If the third case occurs, and the current segment is a `{**}` operator, the algorithm checks if the path ends with any of the suffixes stored in the current node.
The suffix search is performed from the segment that has the same index as the current segment in the checked exact path.
If the suffix is found, the algorithm concludes that the paths conflict.

If the search did not finish in any of the cases above, the algorithm continues until there are no more segments in the checked path.
Then, the algorithm checks if the current node is a possible end of the path. (In the previously presented graph, those nodes are marked with a double circle.)

### Path that contains the `{*}` operator

In case the path that is currently being inserted contains the `{*}` operator,
the algorithm performs the check similarly to the exact path validation.
However, all the children of the current node are considered as possible paths that can be taken.
The algorithm recursively checks all the children of the current node in DFS fashion,
and if any of them contain a valid path, the algorithm concludes that the paths conflict.

### Path that contains the `{**}` operator

In case the path that is currently being inserted contains the `{**}` operator,
the algorithm is interrupted after the `{**}` operator is found (as it must be the last operator in the path).
Then, the algorithm checks if there are any paths that end in the same suffix as the currently checked path.
Example: if the path that is currently being inserted is `b/{**}/a/b`, the algorithm checks if there are any paths after `b` node that end in `/a/b`.
