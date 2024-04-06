package processor

import (
  "context"
  "crypto/aes"
  "crypto/cipher"
  "encoding/base64"
  "encoding/hex"

  "github.com/benthosdev/benthos/v4/public/service"
  "github.com/tidwall/gjson"
  "github.com/tidwall/sjson"
)

func init() {
  configSpec := service.NewConfigSpec().
    Version("0.0.1").
    Summary("Decrypts specific fields with a provided key").
    Field(service.NewStringListField("fields").
      Description("List of fields to decrypt.")).
    Field(service.NewStringField("keyString").
      Description("The key used to decrypt."))

  constructor := func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
    return newDecryptProcessor(conf, mgr.Logger()), nil
  }

  err := service.RegisterProcessor("decrypt", configSpec, constructor)
  if err != nil {
    panic(err)
  }
}

//------------------------------------------------------------------------------

type decryptProcessor struct {
  conf   *service.ParsedConfig
  logger *service.Logger
}

func newDecryptProcessor(conf *service.ParsedConfig, logger *service.Logger) *decryptProcessor {
  return &decryptProcessor{
    conf:   conf,
    logger: logger,
  }
}

func (r *decryptProcessor) Process(ctx context.Context, m *service.Message) (service.MessageBatch, error) {
  bytesContent, err := m.AsBytes()
  if err != nil {
    return nil, err
  }

  keyString, err := r.conf.FieldString("keyString")
  fields, err := r.conf.FieldStringList("fields")
  for _, field := range fields {
    textBytes := gjson.GetBytes(bytesContent, field)
    text := textBytes.String()
    //r.logger.Infof("encrypted text: %s", text)
    cryptoText := decrypt(keyString, text)
    //r.logger.Infof("decrypted text: %s", cryptoText)
    value, _ := sjson.SetBytes(bytesContent, field, cryptoText)
    bytesContent = value
  }
  m.SetBytes(bytesContent)
  return []*service.Message{m}, nil
}

func (r *decryptProcessor) Close(ctx context.Context) error {
  return nil
}

func decrypt(keyString string, encryptedEncodedBytes string) (decryptedString string) {
  key, _ := hex.DecodeString(keyString)
  cipherBytes, _ := base64.URLEncoding.DecodeString(encryptedEncodedBytes)

  block, err := aes.NewCipher(key)
  if err != nil {
    panic(err.Error())
  }

  if len(cipherBytes) < aes.BlockSize {
    panic("ciphertext too short")
  }
  iv := cipherBytes[:aes.BlockSize]
  cipherBytes = cipherBytes[aes.BlockSize:]

  stream := cipher.NewCFBDecrypter(block, iv)
  stream.XORKeyStream(cipherBytes, cipherBytes)

  return string(cipherBytes)
}

