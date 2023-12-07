package payment_gateway
import ("fmt")

const param = "`https://api-beta.sao.network/SaoNetwork/sao/model/metadata/${dataId}`"

const source = ```const dataId= args[0]

const metadata = Functions.makeHttpRequest({
    url : param,
})

const [metadataResp] = await Promise.all([metadata])

let status = 0;

if (metadataResp.data) {
    status = metadataResp.data.metadata.status
}

return Functions.encodeUint256(status)
```

const QueryOrderSource = fmt.Sprintf(source, param)
