# Voice Changer Feature Documentation

## Overview

The Voice Changer feature integrates ElevenLabs Speech-to-Speech API to transform audio between different voices while preserving emotion, delivery, and performance nuances. This allows users to convert their audio recordings to different voices for practice, analysis, and content creation.

## Features

- **Voice Conversion**: Transform any source audio to a different cloned voice
- **Emotion Preservation**: Maintains whispers, laughs, cries, accents, and emotional cues
- **Multilingual Support**: Supports 29 languages via `eleven_multilingual_sts_v2` model
- **Customizable Parameters**: Adjust style, stability, and background noise removal
- **Audio Preview**: Play original and converted audio side-by-side
- **Download Support**: Download converted audio files

## API Endpoint

### POST `/api/v1/voice/convert`

Converts audio to a different voice using ElevenLabs Speech-to-Speech API.

**Authentication**: Required (Bearer token)

**Request Format**: `multipart/form-data`

**Parameters**:
- `audio` (file, required): Audio file to convert (mp3, wav, m4a, etc.)
- `voice_id` (string, required): ElevenLabs voice ID to convert to
- `model_id` (string, optional): Model to use (default: `eleven_multilingual_sts_v2`)
  - `eleven_multilingual_sts_v2`: Supports 29 languages
  - `eleven_english_sts_v2`: English only
- `style` (float, optional): Style parameter 0.0-1.0 (default: 0.0)
  - 0%: Use when input audio is already expressive
  - Higher values add interpretative flair
- `stability` (float, optional): Stability parameter 0.0-1.0 (default: 1.0)
  - 100%: Maximum voice consistency
  - Lower values allow more variation
- `remove_background_noise` (boolean, optional): Remove background noise (default: false)

**Response**:
- **Success (200)**: Audio file (binary, `audio/mpeg`)
- **Error (400/401/500)**: JSON error message

**Example Request**:
```bash
curl -X POST http://localhost:8080/api/v1/voice/convert \
  -H "Authorization: Bearer <token>" \
  -F "audio=@recording.mp3" \
  -F "voice_id=pNInz6obpgDQGcFmaJgB" \
  -F "model_id=eleven_multilingual_sts_v2" \
  -F "style=0.0" \
  -F "stability=1.0" \
  -F "remove_background_noise=false"
```

## Backend Implementation

### Service Layer

**File**: `internal/services/elevenlabs.go`

**Method**: `SpeechToSpeech()`

```go
func (s *ElevenLabsService) SpeechToSpeech(
    voiceID string,
    audioData []byte,
    modelID string,
    style float64,
    stability float64,
    removeBackgroundNoise bool,
) ([]byte, error)
```

**Parameters**:
- `voiceID`: Target voice ID
- `audioData`: Source audio file bytes
- `modelID`: Model identifier
- `style`: Style parameter (0.0-1.0)
- `stability`: Stability parameter (0.0-1.0)
- `removeBackgroundNoise`: Background noise removal flag

**Returns**: Converted audio bytes or error

### Controller Layer

**File**: `api/voice_controller.go`

**Function**: `ConvertVoiceAPI()`

Handles:
1. User authentication
2. Multipart form parsing
3. Parameter validation
4. ElevenLabs API call
5. Audio response streaming

## Frontend Implementation

### Service Layer

**File**: `src/lib/service/voice.service.tsx`

**Function**: `convertVoice()`

```typescript
export async function convertVoice(
  audioFile: File,
  params: VoiceConvertParams
): Promise<Blob>
```

**Interface**:
```typescript
interface VoiceConvertParams {
  voiceId: string;
  modelId?: string;
  style?: number;
  stability?: number;
  removeBackgroundNoise?: boolean;
}
```

### UI Component

**File**: `src/app/voice/page.tsx`

**Features**:
- File upload with drag & drop support
- Voice ID input field
- Model selection
- Style and stability sliders
- Background noise removal toggle
- Audio players for original and converted audio
- Download functionality
- Error handling and loading states

**Route**: `/voice`

## Supported Languages

The `eleven_multilingual_sts_v2` model supports 29 languages:

English (USA, UK, Australia, Canada), Japanese, Chinese, German, Hindi, French (France, Canada), Korean, Portuguese (Brazil, Portugal), Italian, Spanish (Spain, Mexico), Indonesian, Dutch, Turkish, Filipino, Polish, Swedish, Bulgarian, Romanian, Arabic (Saudi Arabia, UAE), Czech, Greek, Finnish, Croatian, Malay, Slovak, Danish, Tamil, Ukrainian & Russian.

## Best Practices

### Audio Quality
- Record in a quiet environment to minimize background noise
- Maintain appropriate microphone levels - avoid too quiet or peaked audio
- Use `remove_background_noise=true` if environmental sounds are present

### Recording Guidelines
- Keep segments under 5 minutes for optimal processing
- Feel free to include natural expressions (laughs, sighs, emotions)
- The source audio's accent and language will be preserved in the output

### Parameters
- **Style**: Set to 0% when input audio is already expressive
- **Stability**: Use 100% for maximum voice consistency
- **Language**: Choose source audio that matches your desired accent and language

## Limitations

1. **File Size**: Audio files should be under 10MB (approximately 5 minutes)
2. **Processing Time**: Conversion may take several seconds depending on file size
3. **API Rate Limits**: Subject to ElevenLabs API rate limits
4. **Audio Format**: Supports common audio formats (mp3, wav, m4a, etc.)

## Environment Variables

Required in `.env`:
```
ELEVENLABS_API_KEY=your_api_key_here
```

## Error Handling

Common errors:
- `401 Unauthorized`: Invalid or missing authentication token
- `400 Bad Request`: Missing required parameters or invalid file format
- `500 Internal Server Error`: ElevenLabs API error or server issue

## Future Enhancements

Potential improvements:
- Voice library browser (fetch available voices from ElevenLabs)
- Audio recording directly in browser
- Batch conversion for multiple files
- Conversion history tracking
- Integration with interview sessions for post-interview review
- Pre-interview practice mode with voice conversion

## Testing

To test the endpoint:

1. Start the backend server:
   ```bash
   go run main.go
   ```

2. Use curl or Postman to send a multipart form request with an audio file

3. Verify the response is an audio file that can be played

## Related Files

- Backend:
  - `internal/services/elevenlabs.go` - Service implementation
  - `api/voice_controller.go` - API controller
  - `main.go` - Route registration

- Frontend:
  - `src/lib/service/voice.service.tsx` - Service layer
  - `src/app/voice/page.tsx` - UI component

