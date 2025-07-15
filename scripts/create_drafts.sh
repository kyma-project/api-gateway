#!/bin/bash

# === CONFIGURATION ===
ISSUE_TITLE=$1
ISSUE_BODY=$2

API_URL="https://api.github.com/graphql"
HEADER="Authorization: bearer $GH_TOKEN"


# === STEP 1: Get Project ID ===
project_id=$(curl -s -H "$HEADER" -X POST $API_URL -d @- <<EOF | jq -r '.data.organization.projectV2.id'
{
  "query": "query {
    organization(login: \"$ORG\") {
      projectV2(number: $PROJECT_NUMBER) {
        id
      }
    }
  }"
}
EOF
)

echo "Project ID: $project_id"
if [[ "$project_id" == "null" || -z "$project_id" ]]; then
  echo "❌ Could not retrieve project ID. Check ORG and PROJECT_NUMBER."
  exit 1
fi

# === STEP 2: Create Draft Issue ===
draft_issue_id=$(curl -s -H "$HEADER" -X POST $API_URL -d @- <<EOF | jq -r '.data.addProjectV2DraftIssue.projectItem.id'
{
  "query": "mutation {
    addProjectV2DraftIssue(input: {
      projectId: \"$project_id\",
      title: \"$ISSUE_TITLE\",
      body: \"$ISSUE_BODY\"
    }) {
      projectItem {
        id
      }
    }
  }"
}
EOF
)

echo "Draft Issue ID: $draft_issue_id"
if [[ "$draft_issue_id" == "null" || -z "$draft_issue_id" ]]; then
  echo "❌ Failed to create draft issue."
  exit 1
fi

# === STEP 3: Get Status Field & Option ID ===
# Fetch all fields and extract Status field ID + To do option ID
read status_field_id status_option_id <<<$(curl -s -H "$HEADER" -X POST $API_URL -d @- <<EOF | jq -r --arg STATUS_NAME "$STATUS_NAME" '
  .data.node.fields.nodes[] 
  | select(.name == "Status") 
  | [.id, (.options[] | select(.name == $STATUS_NAME).id)] 
  | @tsv
'
{
  "query": "query {
    node(id: \"$project_id\") {
      ... on ProjectV2 {
        fields(first: 20) {
          nodes {
            ... on ProjectV2SingleSelectField {
              id
              name
              options {
                id
                name
              }
            }
          }
        }
      }
    }
  }"
}
EOF
)

echo "Status Field ID: $status_field_id"
echo "Option ID for '$STATUS_NAME': $status_option_id"

if [[ -z "$status_field_id" || -z "$status_option_id" ]]; then
  echo "❌ Could not find status field or '$STATUS_NAME' option."
  exit 1
fi
# === STEP 4: Get "Ready" Field & "Sprint" Option ID ===


read ready_field_id ready_option_id <<<$(curl -s -H "$HEADER" -X POST $API_URL -d @- <<EOF | jq -r --arg READY_NAME "$READY_NAME" --arg READY_VALUE "$READY_VALUE" '
  .data.node.fields.nodes[] 
  | select(.name == $READY_NAME) 
  | [.id, (.options[] | select(.name == $READY_VALUE).id)] 
  | @tsv
'
{
  "query": "query {
    node(id: \"$project_id\") {
      ... on ProjectV2 {
        fields(first: 30) {
          nodes {
            ... on ProjectV2SingleSelectField {
              id
              name
              options {
                id
                name
              }
            }
          }
        }
      }
    }
  }"
}
EOF
)

echo "Ready Field ID: $ready_field_id"
echo "Option ID for 'Sprint': $ready_option_id"

if [[ -z "$ready_field_id" || -z "$ready_option_id" ]]; then
  echo "❌ Could not find Ready field or 'Sprint' option."
  exit 1
fi

# === STEP 5: Set Status Field on Draft Issue ===
curl -s -H "$HEADER" -X POST $API_URL -d @- <<EOF | jq
{
  "query": "mutation {
    updateProjectV2ItemFieldValue(input: {
      projectId: \"$project_id\",
      itemId: \"$draft_issue_id\",
      fieldId: \"$status_field_id\",
      value: {
        singleSelectOptionId: \"$status_option_id\"
      }
    }) 
    {
      projectV2Item {
        id
      }


    }
  }"
}
EOF
# === STEP 6: Set Ready Field on Draft Issue ===
curl -s -H "$HEADER" -X POST $API_URL -d @- <<EOF | jq
{
  "query": "mutation {
    updateProjectV2ItemFieldValue(input: {
      projectId: \"$project_id\",
      itemId: \"$draft_issue_id\",
      fieldId: \"$ready_field_id\",
      value: {
        singleSelectOptionId: \"$ready_option_id\"
      }
    }) 
    {
      projectV2Item {
        id
      }


    }
  }"
}
EOF
echo "✅ Draft issue created and status set to '$STATUS_NAME' and '$READY_NAME' "
