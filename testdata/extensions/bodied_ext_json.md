```adf:extension
{
  "type": "bodiedExtension",
  "attrs": {
    "extensionKey": "panel",
    "extensionType": "com.atlassian.confluence.macro.core",
    "parameters": {
      "macroParams": {
        "title": {
          "value": "Next you might want to:"
        }
      }
    }
  },
  "content": [
    {
      "type": "taskList",
      "content": [
        {
          "type": "taskItem",
          "content": [
            {
              "type": "text",
              "text": "Customise the overview page",
              "marks": [
                {
                  "type": "strong"
                }
              ]
            },
            {
              "type": "text",
              "text": " - Click the pencil icon..."
            }
          ],
          "attrs": {
            "localId": "task-1",
            "state": "TODO"
          }
        },
        {
          "type": "taskItem",
          "content": [
            {
              "type": "text",
              "text": "Create additional pages",
              "marks": [
                {
                  "type": "strong"
                }
              ]
            },
            {
              "type": "text",
              "text": " - Click the + in the left sidebar..."
            }
          ],
          "attrs": {
            "localId": "task-2",
            "state": "TODO"
          }
        }
      ]
    }
  ]
}
```
