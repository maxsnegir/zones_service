local json = require("dkjson")

-- init random
math.randomseed(os.time())

-- the request function that will run at each request
request = function()
    local body = {}

    zones_cnt = math.random(1, 100)
    ids = {}
    for j = 1, zones_cnt do
        ids[j] = j
    end

    body["ids"] = ids
    body["point"] = {
        lat = math.random(-90, 90),
        lon = math.random(-180, 180)
    }
    return wrk.format(nil, nil, nil, json.encode(body))
end

--response = function(status, headers, body)
--   print("status: " .. status)
--   print("body: " .. body)
--end

wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"