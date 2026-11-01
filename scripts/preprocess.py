#!/usr/bin/env python3
"""Preprocess a UPS official OpenAPI YAML for oapi-codegen.

Reads --in, writes cleaned YAML to --out. Official specs stay outside git
(.cache/specs). Fixes: tabs, illegal operationIds, servers, OAS 3.1→3.0,
nullable unions, string/bool schema quirks, orphan path parameters.
"""
from __future__ import annotations

import argparse
import re
import sys
from typing import Any

import yaml


def to_pascal(s: str) -> str:
    parts = re.split(r"[^A-Za-z0-9]+", s.strip())
    out = "".join(p[:1].upper() + p[1:] for p in parts if p)
    if not out:
        return "Operation"
    if out[0].isdigit():
        out = "Op" + out
    return out


def fix_operation_ids(doc: dict[str, Any]) -> None:
    """Normalize operationIds and avoid oapi-codegen collisions with schema names.

    oapi-codegen names the client response wrapper `{operationId}Response`, which
    clashes when components.schemas already defines that name (e.g. Rate + RateResponse).
    """
    schemas = set(((doc.get("components") or {}).get("schemas") or {}).keys())
    paths = doc.get("paths") or {}
    used: set[str] = set()
    for path, item in paths.items():
        if not isinstance(item, dict):
            continue
        for method, op in list(item.items()):
            if method in ("parameters", "servers", "summary", "description") or method.startswith("x-"):
                continue
            if not isinstance(op, dict):
                continue
            oid = op.get("operationId")
            if oid is None:
                oid = to_pascal(f"{method}_{path}")
            else:
                # Always PascalCase so collision checks match Go export names /
                # schema identifiers (UPS often uses camelCase operationIds).
                oid = to_pascal(str(oid))
            base = oid
            n = 2
            while (
                oid in used
                or oid in schemas
                or f"{oid}Response" in schemas
                or f"{oid}Request" in schemas
            ):
                if n == 2 and (f"{base}Response" in schemas or base in schemas):
                    oid = f"{base}Op"
                else:
                    oid = f"{base}{n}"
                n += 1
            used.add(oid)
            op["operationId"] = oid
            op.pop("servers", None)


def set_servers(doc: dict[str, Any], cie: str, prod: str) -> None:
    doc["servers"] = [
        {"url": cie, "description": "Customer Integration Environment"},
        {"url": prod, "description": "Production"},
    ]
    paths = doc.get("paths") or {}
    for item in paths.values():
        if isinstance(item, dict):
            item.pop("servers", None)


def path_braces(path: str) -> set[str]:
    return set(re.findall(r"\{([^}]+)\}", path))


def param_key(p: dict[str, Any]) -> tuple[Any, Any]:
    return (p.get("in"), p.get("name"))


def dedupe_params(params: list[Any]) -> list[Any]:
    out: list[Any] = []
    seen: set[tuple[Any, Any]] = set()
    for p in params:
        if not isinstance(p, dict):
            out.append(p)
            continue
        key = param_key(p)
        if key in seen:
            continue
        seen.add(key)
        out.append(p)
    return out


def filter_params(params: Any, braces: set[str]) -> list[Any]:
    if not isinstance(params, list):
        return []
    out = []
    for p in params:
        if not isinstance(p, dict):
            out.append(p)
            continue
        if p.get("in") == "path" and p.get("name") not in braces:
            # Orphan path param (often {version} lived on server URL).
            continue
        out.append(p)
    return dedupe_params(out)


def ensure_path_params(params: list[Any], braces: set[str]) -> list[Any]:
    """Ensure every {brace} has a required path parameter declaration."""
    present = set()
    for p in params:
        if isinstance(p, dict) and p.get("in") == "path":
            present.add(p.get("name"))
    for name in sorted(braces - present):
        params.append(
            {
                "name": name,
                "in": "path",
                "required": True,
                "schema": {"type": "string"},
            }
        )
    return dedupe_params(params)


def fix_path_parameters(doc: dict[str, Any]) -> None:
    paths = doc.get("paths") or {}
    for path, item in paths.items():
        if not isinstance(item, dict):
            continue
        braces = path_braces(path)
        path_level_all = item.get("parameters") or []
        if not isinstance(path_level_all, list):
            path_level_all = []
        path_level_path = [p for p in path_level_all if isinstance(p, dict) and p.get("in") == "path"]
        path_level_other = [p for p in path_level_all if not (isinstance(p, dict) and p.get("in") == "path")]

        for method, op in item.items():
            if not isinstance(op, dict) or method in ("parameters", "servers"):
                continue
            existing = op.get("parameters") or []
            if not isinstance(existing, list):
                existing = []
            # Fold path-level path params into the operation (once), then dedupe.
            merged = filter_params(list(existing) + path_level_path, braces)
            op["parameters"] = ensure_path_params(merged, braces)

        # Keep only non-path params at path item level to avoid duplicate path/* with operations.
        if path_level_other or "parameters" in item:
            item["parameters"] = filter_params(path_level_other, braces)
            if not item["parameters"]:
                item.pop("parameters", None)


def simplify_deep_property_refs(doc: dict[str, Any]) -> None:
    """Replace #/components/schemas/X/properties/Y refs with inline object schemas."""
    ref_re = re.compile(r"^#/components/schemas/([^/]+)/properties/(.+)$")

    def resolve(ref: str) -> Any:
        m = ref_re.match(ref)
        if not m:
            return None
        schema_name, prop_path = m.group(1), m.group(2).split("/")
        schemas = (doc.get("components") or {}).get("schemas") or {}
        cur: Any = schemas.get(schema_name)
        for part in prop_path:
            if not isinstance(cur, dict):
                return {"type": "object", "additionalProperties": True}
            if part == "properties":
                continue
            props = cur.get("properties") if isinstance(cur.get("properties"), dict) else None
            if props is not None and part in props:
                cur = props[part]
            elif part in cur:
                cur = cur[part]
            else:
                return {"type": "object", "additionalProperties": True}
        if isinstance(cur, dict):
            # Shallow copy without nested $ref loops
            out = {k: v for k, v in cur.items() if k != "$ref"}
            return out or {"type": "object", "additionalProperties": True}
        return {"type": "object", "additionalProperties": True}

    def walk(o: Any) -> Any:
        if isinstance(o, list):
            return [walk(x) for x in o]
        if not isinstance(o, dict):
            return o
        if "$ref" in o and isinstance(o["$ref"], str) and "/properties/" in o["$ref"]:
            resolved = resolve(o["$ref"])
            if resolved is not None:
                return walk(resolved)
        return {k: walk(v) for k, v in o.items()}

    if "components" in doc:
        doc["components"] = walk(doc["components"])
    if "paths" in doc:
        doc["paths"] = walk(doc["paths"])


def normalize_type_list(node: dict[str, Any]) -> None:
    t = node.get("type")
    if isinstance(t, list):
        non_null = [x for x in t if x != "null"]
        if "null" in t:
            node["nullable"] = True
        if len(non_null) == 1:
            node["type"] = non_null[0]
        elif len(non_null) == 0:
            node.pop("type", None)
            node["nullable"] = True
        else:
            # multi-type without a single primitive — fall back to string
            node["type"] = non_null[0]
    if node.get("type") == "null":
        node.pop("type", None)
        node["nullable"] = True


def fix_schema_node(node: Any) -> Any:
    if isinstance(node, list):
        return [fix_schema_node(x) for x in node]
    if not isinstance(node, dict):
        return node

    # required must be a list of strings
    if "required" in node and isinstance(node["required"], str):
        node["required"] = [node["required"]]
    if "required" in node and isinstance(node["required"], list):
        node["required"] = [x for x in node["required"] if isinstance(x, str)]

    # properties must be a map
    if "properties" in node and not isinstance(node["properties"], dict):
        node.pop("properties", None)

    # UPS bug: additionalProperties / $schema / xml sometimes nested inside properties as bool/scalar.
    props = node.get("properties")
    if isinstance(props, dict):
        for bad_key in ("additionalProperties", "unevaluatedProperties", "$schema"):
            if bad_key in props and not isinstance(props[bad_key], dict):
                val = props.pop(bad_key)
                if bad_key not in node:
                    node[bad_key] = val
        # Drop any property whose schema is a bare scalar/bool (invalid).
        for pk, pv in list(props.items()):
            if not isinstance(pv, (dict, list)):
                props.pop(pk)

    # items must be a schema object (or bool in JSON Schema — kin-openapi hates bool)
    if "items" in node and isinstance(node["items"], bool):
        node["items"] = {} if node["items"] else {"not": {}}

    normalize_type_list(node)

    # Unsupported / awkward formats for oapi-codegen
    fmt = node.get("format")
    if fmt == "decimal":
        node["format"] = "double"
    elif fmt == "integer" and node.get("type") == "number":
        node["type"] = "integer"
        node.pop("format", None)
    elif fmt == "integer" and node.get("type") not in ("integer", None):
        node.pop("format", None)

    # OAS 3.1 exclusiveMinimum as number → 3.0 boolean + minimum
    for excl, soft in (("exclusiveMinimum", "minimum"), ("exclusiveMaximum", "maximum")):
        if excl in node and not isinstance(node[excl], bool):
            node[soft] = node[excl]
            node[excl] = True

    # oneOf of const-only schemas → enum (before const rewrite on children)
    one_of = node.get("oneOf")
    if isinstance(one_of, list) and one_of and all(
        isinstance(x, dict) and "const" in x and set(x.keys()) <= {"title", "description", "const", "type"}
        for x in one_of
    ):
        node["enum"] = [x["const"] for x in one_of]
        node.setdefault("type", "string")
        node.pop("oneOf", None)

    # const (3.1) → enum for 3.0 / oapi-codegen
    if "const" in node:
        node.setdefault("enum", [node["const"]])
        if "type" not in node:
            c = node["const"]
            if isinstance(c, bool):
                node["type"] = "boolean"
            elif isinstance(c, (int, float)) and not isinstance(c, bool):
                node["type"] = "number"
            else:
                node["type"] = "string"
        node.pop("const", None)

    # Stringify date examples so dump/load stays stable
    if "example" in node and hasattr(node["example"], "isoformat"):
        node["example"] = node["example"].isoformat()

    for k, v in list(node.items()):
        node[k] = fix_schema_node(v)
    return node


def downgrade_openapi(doc: dict[str, Any]) -> None:
    ver = str(doc.get("openapi") or "")
    if ver.startswith("3.1"):
        doc["openapi"] = "3.0.3"
    fix_schema_node(doc)


def load_yaml(raw: str) -> dict[str, Any]:
    cleaned = raw.replace("\t", "  ")
    doc = yaml.safe_load(cleaned)
    if not isinstance(doc, dict):
        raise SystemExit("expected OpenAPI document mapping")
    return doc


def patch_ups_spec_quirks(doc: dict[str, Any]) -> None:
    """Fix official UPS OpenAPI mismatches with live JSON responses."""
    schemas = ((doc.get("components") or {}).get("schemas") or {})
    if not isinstance(schemas, dict):
        return

    # Rating: NegotiatedRateCharges.BaseServiceCharge is documented as an array,
    # but CIE/prod return a single object (same shape as RatedShipment.BaseServiceCharge).
    nrc = schemas.get("RatedShipment_NegotiatedRateCharges")
    if isinstance(nrc, dict):
        props = nrc.get("properties")
        if isinstance(props, dict):
            bsc = props.get("BaseServiceCharge")
            if isinstance(bsc, dict) and bsc.get("type") == "array":
                items = bsc.get("items")
                ref = items.get("$ref") if isinstance(items, dict) else None
                if ref:
                    props["BaseServiceCharge"] = {"$ref": ref}


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--in", dest="infile", required=True)
    ap.add_argument("--out", dest="outfile", required=True)
    ap.add_argument("--server-cie", required=True)
    ap.add_argument("--server-prod", required=True)
    args = ap.parse_args()

    raw = open(args.infile, encoding="utf-8").read()
    doc = load_yaml(raw)
    if "openapi" not in doc:
        raise SystemExit("missing openapi version")
    downgrade_openapi(doc)
    fix_operation_ids(doc)
    set_servers(doc, args.server_cie, args.server_prod)
    fix_path_parameters(doc)
    simplify_deep_property_refs(doc)
    patch_ups_spec_quirks(doc)

    with open(args.outfile, "w", encoding="utf-8") as f:
        yaml.safe_dump(
            doc,
            f,
            sort_keys=False,
            allow_unicode=True,
            width=120,
            default_flow_style=False,
        )


if __name__ == "__main__":
    main()
