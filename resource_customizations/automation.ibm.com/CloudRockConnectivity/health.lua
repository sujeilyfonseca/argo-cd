hs = {}

if obj.status ~= nil then
  if obj.status.conditions ~= nil then
    for i, condition in ipairs(obj.status.conditions) do
      if condition.type == "Ready" and condition.status == "False" then
        hs.status = "Degraded"
        if obj.status.message ~= nil
          then hs.message = obj.status.message
          else hs.message = "CloudRockConnectivity degraded"
        end
        return hs
      end
      if condition.type == "Ready" and condition.status == "True" then
        hs.status = "Healthy"
        if obj.status.message ~= nil
          then  hs.message = obj.status.message
          else hs.message = "CloudRockConnectivity healthy"
        end
        return hs
      end
    end
  end
end

hs.status = "Progressing"
if obj.status.message ~= nil
  then hs.message = obj.status.message
  else hs.message = "CloudRockConnectivity progressing"
end
return hs
