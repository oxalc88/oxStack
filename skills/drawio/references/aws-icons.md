# AWS Icons Reference (image-based icon set)

This reference describes the official AWS Architecture Icons layout for projects that vendor the actual image files into their own repo. **The image files themselves are not bundled with this skill** (the full set is ~22MB across 4000+ PNG/SVG files) — only this lookup table travels with the skill. See the "AWS DIAGRAMS" section in `SKILL.md` for when to use this vs. the native `mxgraph.aws4` shapes.

If a project needs these icons, copy the relevant folders from the official AWS Architecture Icons package into that project (e.g. under `.claude/skills/drawio/icons/` or wherever that repo keeps vendored assets) and adjust the paths below to match.

## Expected layout once vendored

| Set | Use for | Path pattern |
|-----|---------|--------------|
| `Architecture-Service-Icons_<date>/` | AWS services | `Arch_<Category>/<size>/Arch_<ServiceName>_<size>.svg` |
| `Resource-Icons_<date>/` | Sub-resources within a service | `Res_<Category>/Res_<ServiceName>_<Resource>_48.svg` |
| `Category-Icons_<date>/` | Domain grouping labels | `Arch-Category_<size>/Arch-Category_<Name>_<size>.svg` |
| `Architecture-Group-Icons_<date>/` | Boundary containers (VPC, Region…) | `<GroupName>_32.svg` (flat, 32px only) |

Service and Category icons come in sizes 16/32/48/64. Resource icons are 48 only. **Use 48 for everything.**

To find an icon, list the matching category folder and pick the closest filename. Prefer `.svg` over `.png`.

## Embedding in XML

```xml
<mxCell id="2" value="Amazon EC2"
  style="shape=image;image=<path-to-vendored-icons>/Architecture-Service-Icons_<date>/Arch_Compute/48/Arch_Amazon-EC2_48.svg;labelBackgroundColor=none;labelPosition=center;verticalLabelPosition=bottom;verticalAlign=top;fillColor=none;strokeColor=none;"
  vertex="1" parent="1">
  <mxGeometry x="160" y="160" width="48" height="48" as="geometry"/>
</mxCell>
```

Replace `<path-to-vendored-icons>` with the actual relative path from the `.drawio` file to wherever the project vendored the icon set (paths are project-specific — this skill install does not ship the files).

## Common services quick reference

| Service | Category | File |
|---------|----------|------|
| EC2 | `Arch_Compute` | `Arch_Amazon-EC2_48.svg` |
| S3 | `Arch_Storage` | `Arch_Amazon-Simple-Storage-Service_48.svg` |
| Lambda | `Arch_Compute` | `Arch_AWS-Lambda_48.svg` |
| RDS | `Arch_Databases` | `Arch_Amazon-RDS_48.svg` |
| DynamoDB | `Arch_Databases` | `Arch_Amazon-DynamoDB_48.svg` |
| API Gateway | `Arch_Networking-Content-Delivery` | `Arch_Amazon-API-Gateway_48.svg` |
| CloudFront | `Arch_Networking-Content-Delivery` | `Arch_Amazon-CloudFront_48.svg` |
| VPC | `Arch_Networking-Content-Delivery` | `Arch_Amazon-Virtual-Private-Cloud_48.svg` |
| IAM | `Arch_Security-Identity` | `Arch_AWS-Identity-and-Access-Management_48.svg` |
| Cognito | `Arch_Security-Identity` | `Arch_Amazon-Cognito_48.svg` |
| SQS | `Arch_Application-Integration` | `Arch_Amazon-Simple-Queue-Service_48.svg` |
| SNS | `Arch_Application-Integration` | `Arch_Amazon-Simple-Notification-Service_48.svg` |
| ECS | `Arch_Containers` | `Arch_Amazon-Elastic-Container-Service_48.svg` |
| EKS | `Arch_Containers` | `Arch_Amazon-Elastic-Kubernetes-Service_48.svg` |
| CloudWatch | `Arch_Management-Tools` | `Arch_Amazon-CloudWatch_48.svg` |
| Bedrock | `Arch_Artificial-Intelligence` | `Arch_Amazon-Bedrock_48.svg` |
