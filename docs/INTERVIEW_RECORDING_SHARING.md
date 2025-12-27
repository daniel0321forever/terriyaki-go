# Interview Recording & Sharing Feature

## Overview
Enable users to record their coding interviews, optionally share them with their grind groups, and use voice conversion for anonymity. This makes interview practice a social learning experience.

---

## 🎯 What We're Building

### Core Features

1. **Automatic Audio Recording During Interview**
   - Record user's microphone audio throughout the interview session
   - Store audio locally in browser, then upload after interview ends
   - Include both user responses and interviewer questions (full conversation)

2. **Post-Interview Sharing Decision**
   - After interview ends, show dialog: "Share this interview with your grind?"
   - Options: Share with grind / Keep private / Share later
   - If sharing: Apply voice conversion for anonymity

3. **Voice Conversion for Anonymity**
   - Convert user's voice to a neutral/professional voice
   - Apply noise reduction automatically
   - Preserve emotional delivery and speech patterns
   - Hide user's identity while maintaining authenticity

4. **Shared Interview Feed in Grinds**
   - Display shared interviews in grind progress/feed
   - Show: Task name, difficulty, anonymized audio, transcript preview
   - Allow grind members to listen and provide feedback
   - Show who shared (but with anonymized voice)

5. **Interview Archive & Management**
   - View all user's interviews (shared and private)
   - Re-share previously private interviews
   - Delete shared interviews (remove from grind feed)

---

## 🏗️ Technical Architecture

### Data Model Changes

#### 1. Update `InterviewSession` Model
```go
type InterviewSession struct {
    // ... existing fields ...
    
    // New fields for recording
    AudioRecordingURL      string    `json:"audio_recording_url" gorm:""`           // URL to stored audio file
    IsShared               bool      `json:"is_shared" gorm:"default:false"`      // Whether shared to grind
    SharedGrindID          string    `json:"shared_grind_id" gorm:""`            // Grind it's shared to (if shared)
    AnonymizedAudioURL     string    `json:"anonymized_audio_url" gorm:""`       // URL to voice-converted audio
    VoiceConversionApplied bool      `json:"voice_conversion_applied" gorm:"default:false"` // Whether voice was converted
    AnonymizedVoiceID      string    `json:"anonymized_voice_id" gorm:""`         // Voice ID used for conversion
    SharedAt               *time.Time `json:"shared_at" gorm:""`                  // When it was shared
}
```

#### 2. New `SharedInterview` Model (Optional - for feed/notifications)
```go
type SharedInterview struct {
    gorm.Model
    ID              string    `json:"id" gorm:"primaryKey"`
    InterviewID     string    `json:"interview_id" gorm:"not null"`     // FK to InterviewSession
    GrindID         string    `json:"grind_id" gorm:"not null"`        // Grind it's shared to
    SharedBy        string    `json:"shared_by" gorm:"not null"`        // User who shared
    CreatedAt       time.Time `json:"created_at" gorm:"not null"`
}
```

### API Endpoints

#### 1. Upload Interview Audio
```
POST /api/v1/interviews/:id/audio
- Upload audio file after interview ends
- Store in cloud storage (S3/local storage)
- Update InterviewSession with audio URL
```

#### 2. Share Interview
```
POST /api/v1/interviews/:id/share
Body: {
  "grind_id": "string",
  "anonymize": true,  // Always true for now
  "voice_id": "string" // Optional: specific voice for conversion
}
- Convert voice if anonymize=true
- Update InterviewSession.is_shared = true
- Create SharedInterview record
- Return shared interview data
```

#### 3. Get Shared Interviews for Grind
```
GET /api/v1/grinds/:id/shared-interviews
- Get all shared interviews for a grind
- Return with anonymized audio URLs
- Include task details, evaluation summary
```

#### 4. Unshare Interview
```
DELETE /api/v1/interviews/:id/share
- Remove from shared feed
- Update InterviewSession.is_shared = false
- Optionally delete SharedInterview record
```

#### 5. Get User's Interview History
```
GET /api/v1/interviews/history
- Get all user's interviews (shared and private)
- Include sharing status
```

### Frontend Components

#### 1. Audio Recorder Hook
```typescript
// hooks/useInterviewRecorder.ts
- Use MediaRecorder API
- Record during interview session
- Store chunks in memory
- Convert to blob after interview ends
```

#### 2. Share Interview Dialog
```typescript
// components/ShareInterviewDialog.tsx
- Show after interview ends
- Options: Share / Keep Private / Share Later
- If sharing: Show voice selection (optional)
- Upload audio + trigger voice conversion
```

#### 3. Shared Interview Card
```typescript
// components/SharedInterviewCard.tsx
- Display in grind feed
- Show: Task name, difficulty, audio player
- Play anonymized audio
- Show transcript preview
- Show who shared (username, not voice)
```

#### 4. Interview History Page
```typescript
// app/interviews/history/page.tsx
- List all user's interviews
- Show sharing status
- Allow re-sharing private interviews
- Delete shared interviews
```

---

## 🔄 User Flow

### Recording Flow
1. User starts interview → Audio recording begins automatically
2. During interview → Audio chunks stored in browser memory
3. Interview ends → Recording stops
4. Audio blob created → Ready for upload

### Sharing Flow
1. Interview ends → Show results + "Share Interview?" dialog
2. User clicks "Share with Grind" → 
   - Upload original audio to server
   - Server converts voice (background job)
   - Create SharedInterview record
   - Show success message
3. Interview appears in grind feed → Other members can listen

### Listening Flow (Grind Members)
1. View grind feed → See shared interviews
2. Click interview card → 
   - See task details
   - Play anonymized audio
   - View transcript
   - See evaluation summary
3. Provide feedback (optional future feature)

---

## 🛠️ Implementation Steps

### Phase 1: Audio Recording (Frontend)
1. Create `useInterviewRecorder` hook
   - Initialize MediaRecorder when interview starts
   - Record audio chunks during session
   - Stop recording when interview ends
   - Convert chunks to blob

2. Integrate into InterviewPage
   - Start recording when `connectToAgent()` is called
   - Stop recording when `endInterviewSession()` is called
   - Store audio blob in state

### Phase 2: Audio Upload (Backend)
1. Create upload endpoint
   - `POST /api/v1/interviews/:id/audio`
   - Accept multipart audio file
   - Store in file system or cloud storage
   - Update InterviewSession with URL

2. Update InterviewSession model
   - Add audio recording fields
   - Run migration

3. Frontend: Upload audio after interview ends
   - Call upload endpoint
   - Show progress indicator

### Phase 3: Voice Conversion Integration
1. Update share endpoint to use voice conversion
   - When sharing, automatically convert voice
   - Use `remove_background_noise=true`
   - Use neutral/professional voice ID
   - Store converted audio URL

2. Background job (optional)
   - Convert voice asynchronously
   - Notify user when conversion complete

### Phase 4: Sharing UI
1. Create ShareInterviewDialog component
   - Show after interview ends
   - Options: Share / Keep Private
   - If sharing: Select grind (if user in multiple grinds)

2. Update InterviewPage
   - Show dialog after results displayed
   - Handle share action

### Phase 5: Grind Feed Integration
1. Create SharedInterview model (optional)
   - Or use InterviewSession.is_shared flag

2. Create API endpoint
   - `GET /api/v1/grinds/:id/shared-interviews`
   - Return shared interviews for grind

3. Create SharedInterviewCard component
   - Display in grind feed
   - Audio player for anonymized audio
   - Task details, transcript preview

4. Add to Grind Progress page
   - New section: "Shared Interviews"
   - List recent shared interviews

### Phase 6: Interview History
1. Create history endpoint
   - `GET /api/v1/interviews/history`
   - Return user's all interviews

2. Create history page
   - List interviews
   - Show sharing status
   - Allow re-sharing or unsharing

---

## 📋 Technical Details

### Audio Recording (Browser)
```typescript
// Use MediaRecorder API
const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
const mediaRecorder = new MediaRecorder(stream, {
  mimeType: 'audio/webm' // or 'audio/mp4'
});

const chunks: Blob[] = [];
mediaRecorder.ondataavailable = (e) => chunks.push(e.data);
mediaRecorder.onstop = () => {
  const blob = new Blob(chunks, { type: 'audio/webm' });
  // Upload blob
};
```

### Voice Conversion Parameters
```go
// Use consistent voice for anonymity
anonymizedVoiceID := "pNInz6obpgDQGcFmaJgB" // Adam - neutral professional voice
// Or use a dedicated "anonymous" voice

params := VoiceConvertParams{
    VoiceID: anonymizedVoiceID,
    ModelID: "eleven_multilingual_sts_v2",
    Style: 0.0,  // Preserve original style
    Stability: 1.0,  // Maximum consistency
    RemoveBackgroundNoise: true,  // Clean audio
}
```

### Storage Strategy
**Option A: Local File System (MVP)**
- Store audio files in `uploads/interviews/`
- Store URLs in database
- Simple, no external dependencies

**Option B: Cloud Storage (Production)**
- Use S3, Google Cloud Storage, etc.
- Better scalability
- CDN for faster delivery

**Recommendation**: Start with Option A, migrate to Option B later

### File Size Considerations
- Average interview: 3-5 minutes
- Audio at 128kbps: ~3-5 MB per interview
- Need to handle:
  - Large files (>10MB)
  - Multiple interviews per user
  - Storage cleanup (old interviews)

---

## 🎨 UI/UX Design

### Share Dialog
```
┌─────────────────────────────────────┐
│  Share Your Interview?              │
├─────────────────────────────────────┤
│                                     │
│  🎤 Interview completed!            │
│                                     │
│  Would you like to share this       │
│  interview with your grind group?   │
│                                     │
│  ✓ Your voice will be anonymized    │
│  ✓ Background noise removed          │
│  ✓ Only your grind can see it       │
│                                     │
│  [Share with Grind]  [Keep Private] │
│                                     │
└─────────────────────────────────────┘
```

### Shared Interview Card
```
┌─────────────────────────────────────┐
│  Two Sum Problem                    │
│  Medium • Shared by @username        │
├─────────────────────────────────────┤
│  🎧 [Play Audio]                    │
│                                     │
│  Transcript: "I'll use a hash map..."│
│                                     │
│  Score: 85/100                       │
└─────────────────────────────────────┘
```

---

## 🔒 Privacy & Security

1. **Anonymization**
   - Always convert voice when sharing
   - Use consistent voice ID (same for all users)
   - Remove identifying information from audio

2. **Access Control**
   - Only grind members can see shared interviews
   - Users can unshare at any time
   - Private interviews never visible to others

3. **Data Retention**
   - Allow users to delete shared interviews
   - Clean up old audio files (optional: after 30 days)

---

## 🚀 Future Enhancements

1. **Feedback System**
   - Allow grind members to comment on shared interviews
   - Rate interview performance
   - Suggest improvements

2. **Interview Analytics**
   - Track improvement over time
   - Compare with grind average
   - Show progress charts

3. **Voice Selection**
   - Let users choose their anonymized voice
   - Different voices for different moods/styles

4. **Batch Sharing**
   - Share multiple interviews at once
   - Create "interview playlists"

5. **Interview Challenges**
   - Challenge grind members to solve same problem
   - Compare approaches and solutions

---

## 📝 Implementation Checklist

### Backend
- [ ] Update InterviewSession model with audio fields
- [ ] Create migration for new fields
- [ ] Create audio upload endpoint
- [ ] Create share interview endpoint
- [ ] Integrate voice conversion in share flow
- [ ] Create get shared interviews endpoint
- [ ] Create unshare endpoint
- [ ] Create interview history endpoint

### Frontend
- [ ] Create useInterviewRecorder hook
- [ ] Integrate recording into InterviewPage
- [ ] Create ShareInterviewDialog component
- [ ] Create SharedInterviewCard component
- [ ] Add shared interviews to grind feed
- [ ] Create interview history page
- [ ] Add audio upload after interview ends

### Testing
- [ ] Test audio recording in different browsers
- [ ] Test voice conversion quality
- [ ] Test sharing/unsharing flow
- [ ] Test privacy (only grind members see shared)
- [ ] Test file size limits

---

## 🎯 Success Metrics

1. **Adoption**
   - % of interviews that are shared
   - Number of shared interviews per grind

2. **Engagement**
   - Average listens per shared interview
   - Time spent listening to shared interviews

3. **Learning**
   - User feedback on helpfulness
   - Improvement in interview scores over time

---

## ❓ Open Questions

1. **Storage**: Local filesystem or cloud storage?
2. **Voice Selection**: Fixed anonymous voice or user choice?
3. **Retention**: How long to keep audio files?
4. **File Format**: WebM, MP3, or both?
5. **Conversion Timing**: Real-time or background job?
6. **Feedback**: Should grind members be able to comment?

---
