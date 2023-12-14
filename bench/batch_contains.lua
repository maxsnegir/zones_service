local json = require("dkjson")

-- init random
math.randomseed(os.time())

-- the request function that will run at each request
request = function()
    keys_count = math.random(1, 100)
    zone_id = math.random(1, 50)
    local body = {}

    for i = 1, keys_count do
        zones_cnt = math.random(1, 100)
        ids = {}
        for j = 1, zones_cnt do
            ids[j] = j
        end
        table.insert(body, {
            key = string.format("%d", i),
            ids = ids,
            point = {
                lat = math.random(-90, 90),
                lon = math.random(-180, 180)
            }
        })

    end
    return wrk.format(nil, nil, nil, json.encode(body))
end

 --response = function(status, headers, body)
 --   print("status: " .. status)
 --   print("body: " .. body)
 --end

wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"