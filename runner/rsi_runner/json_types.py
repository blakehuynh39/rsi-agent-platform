from __future__ import annotations

from typing import Dict, List, Literal, TypedDict, Union

JsonPrimitive = Union[str, int, float, bool, None]
JsonValue = Union[JsonPrimitive, Dict[str, "JsonValue"], List["JsonValue"]]
JsonObject = Dict[str, JsonValue]


class JsonToolParameters(TypedDict, total=False):
    type: str
    properties: JsonObject
    required: list[str]


class JsonToolFunctionSchema(TypedDict):
    name: str
    description: str
    parameters: JsonToolParameters


class JsonToolWrapperSchema(TypedDict):
    type: Literal["function"]
    function: JsonToolFunctionSchema
